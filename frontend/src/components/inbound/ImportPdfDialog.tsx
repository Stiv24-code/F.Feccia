import { useEffect, useMemo, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { FileText, Check, Info } from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import type {
  DtoInboundOrderRequest,
  DtoInboundOrderResponse,
  DtoPdfExtractedValueDTO,
  DtoPdfRenderPageDTO,
  DtoPdfTemplateFieldDTO,
  DtoPdfTemplateResponse,
} from '@/api/data-contracts';
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import PdfFieldOverlay from './PdfFieldOverlay';
import { getInboundApiError } from './constants';

const DRAFT_KEYS: Array<[keyof DraftFields, string]> = [
  ['client', 'Cliente'], ['sender_email', 'E-mail mittente'],
  ['ref', 'Riferimento'], ['product', 'Prodotto'],
  ['kg', 'Kg'], ['rate', 'Nolo'],
  ['load_date', 'Data carico'], ['load_place', 'Luogo carico'],
  ['delivery_date', 'Data consegna'], ['delivery_place', 'Luogo consegna'],
];

// Sotto questa soglia un valore estratto è considerato "non presente nel
// PDF" (vuoto o letto con poca certezza) — stessa soglia usata per il
// conteggio aggregato "N zone incerte".
const CONFIDENCE_THRESHOLD = 0.65;

interface DraftFields {
  client: string;
  sender_email: string;
  ref: string;
  product: string;
  kg: string;
  rate: string;
  load_date: string;
  load_place: string;
  delivery_date: string;
  delivery_place: string;
  notes: string;
}

const emptyDraft: DraftFields = {
  client: '', sender_email: '', ref: '', product: '', kg: '', rate: '',
  load_date: '', load_place: '', delivery_date: '', delivery_place: '', notes: '',
};

interface ImportPdfDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orders: DtoInboundOrderResponse[];
  onImported?: () => void;
}

// Import di un ordine da PDF in due passi (porting OrderMesh):
// 1) file + mittente (template preselezionato dal mittente, o scelto a mano)
// 2) bozza estratta, correggibile, poi «Crea ordine».
export default function ImportPdfDialog({ open, onOpenChange, orders, onImported }: ImportPdfDialogProps) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [templates, setTemplates] = useState<DtoPdfTemplateResponse[]>([]);
  const [file, setFile] = useState<File | null>(null);
  const [sender, setSender] = useState('');
  const [tplId, setTplId] = useState('');
  const [match, setMatch] = useState<{ found: boolean; name?: string } | null>(null);
  const [step, setStep] = useState<1 | 2>(1);
  const [draft, setDraft] = useState<DraftFields>(emptyDraft);
  const [templateId, setTemplateId] = useState('');
  const [templateFields, setTemplateFields] = useState<DtoPdfTemplateFieldDTO[]>([]);
  const [extraction, setExtraction] = useState({ tplName: '', n: 0, low: 0 });
  const [extractionMeta, setExtractionMeta] = useState<Record<string, DtoPdfExtractedValueDTO>>({});
  const [pages, setPages] = useState<DtoPdfRenderPageDTO[]>([]);
  const [curPage, setCurPage] = useState(0);
  const [busy, setBusy] = useState(false);

  // reset a ogni apertura + carica template
  useEffect(() => {
    if (!open) return;
    setFile(null); setSender(''); setTplId(''); setMatch(null); setStep(1);
    setDraft(emptyDraft); setTemplateId(''); setTemplateFields([]);
    setExtractionMeta({}); setPages([]); setCurPage(0);
    apiClient.v1PdfTemplatesList()
      .then((res) => setTemplates(res.data ?? []))
      .catch((err) => toast.error(getInboundApiError(err)));
  }, [open]);

  // suggerimenti mittente da template e ordini esistenti
  const senderOptions = useMemo(() => {
    const s = new Set<string>();
    templates.forEach((t) => (t.senders ?? []).forEach((x) => { if (!x.startsWith('@')) s.add(x); }));
    orders.forEach((o) => { if (o.sender_email) s.add(o.sender_email.toLowerCase()); });
    return [...s].sort();
  }, [templates, orders]);

  // preselezione template dal mittente
  useEffect(() => {
    if (!open || !sender.includes('@')) { setMatch(null); return; }
    let cancelled = false;
    apiClient.v1PdfTemplatesMatchList({ sender })
      .then((res) => {
        if (cancelled) return;
        const m = res.data?.match;
        if (m?.id) { setTplId(m.id); setMatch({ found: true, name: m.name }); }
        else setMatch({ found: false });
      })
      .catch(() => setMatch(null));
    return () => { cancelled = true; };
  }, [sender, open]);

  async function extract() {
    if (!file) { toast.error('Scegli il PDF dell’ordine'); return; }
    if (!tplId && !sender) { toast.error('Indica il mittente o scegli un template'); return; }
    setBusy(true);
    try {
      const [res, renderRes] = await Promise.all([
        apiClient.v1PdfImportCreate({
          file,
          template_id: tplId || undefined,
          sender: sender || undefined,
        }),
        // La preview del documento è un extra per la revisione: se il
        // render fallisce non deve bloccare l'estrazione dei campi.
        apiClient.v1PdfRenderCreate({ file }).catch(() => null),
      ]);
      const j = res.data;
      const order = j.order ?? {};
      const d = { ...emptyDraft };
      DRAFT_KEYS.forEach(([k]) => { d[k] = String((order as Record<string, unknown>)[k] ?? ''); });
      d.notes = order.notes ?? '';
      const extractionMap = j.extraction ?? {};
      const ex = Object.values(extractionMap);
      setExtraction({
        tplName: j.template?.name ?? '',
        n: ex.filter((v) => (v.confidence ?? 0) >= CONFIDENCE_THRESHOLD).length,
        low: ex.filter((v) => (v.confidence ?? 0) < CONFIDENCE_THRESHOLD).length,
      });
      setExtractionMeta(extractionMap);
      setDraft(d);
      const usedTemplateId = j.template?.id || tplId;
      setTemplateId(usedTemplateId ?? '');
      setTemplateFields(templates.find((t) => t.id === usedTemplateId)?.fields ?? []);
      setPages(renderRes?.data.pages ?? []);
      setCurPage(0);
      setStep(2);
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy(false);
    }
  }

  // Target dei campi non letti dal PDF o letti con poca certezza — pilotano
  // sia il badge per campo nel form, sia l'evidenziazione nel documento.
  const missingTargets = useMemo(() => {
    const s = new Set<string>();
    DRAFT_KEYS.forEach(([k]) => {
      const meta = extractionMeta[k];
      if (!meta || (meta.confidence ?? 0) < CONFIDENCE_THRESHOLD) s.add(k);
    });
    return s;
  }, [extractionMeta]);

  const currentPage = pages.find((p) => p.page_num === curPage) ?? pages[0] ?? null;
  const overlayFields = templateFields.filter((f) => (f.page ?? 0) === (currentPage?.page_num ?? 0));

  async function save() {
    const body: DtoInboundOrderRequest = {
      client: draft.client.trim(),
      sender_email: draft.sender_email.trim(),
      ref: draft.ref.trim(),
      product: draft.product.trim(),
      kg: parseInt(draft.kg.replace(/[^\d]/g, ''), 10) || 0,
      rate: draft.rate.trim(),
      load_date: draft.load_date.trim(),
      load_place: draft.load_place.trim(),
      delivery_date: draft.delivery_date.trim(),
      delivery_place: draft.delivery_place.trim(),
      notes: draft.notes.trim(),
      source: 'pdf',
      template_id: templateId || undefined,
    };
    setBusy(true);
    try {
      await apiClient.v1InboundOrdersCreate(body);
      onOpenChange(false);
      toast.success(`Ordine ${body.ref} importato dal PDF (template ${extraction.tplName})`);
      onImported?.();
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={step === 2 ? 'sm:max-w-[1100px] w-[95vw] max-h-[85vh] overflow-y-auto' : 'sm:max-w-[560px]'}>
        <DialogHeader>
          <DialogTitle>{step === 2 ? 'Revisione ordine da PDF' : 'Importa ordine da PDF'}</DialogTitle>
          <DialogDescription>
            {step === 2 ? (
              <>
                Correggi dove serve: alla conferma l’ordine entra nel Registro come «Da
                confermare».{' '}
                <Link to="/ordini-in-ingresso/template" className="underline" onClick={() => onOpenChange(false)}>
                  Gestisci i template PDF →
                </Link>
              </>
            ) : (
              <>
                Il template giusto viene preselezionato in base al mittente; puoi comunque
                sceglierlo a mano. I template si creano in{' '}
                <Link to="/ordini-in-ingresso/template" className="underline" onClick={() => onOpenChange(false)}>
                  Template PDF
                </Link>.
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        {step === 1 && (
          <div className="flex flex-col gap-3">
            <input
              ref={fileRef} type="file" accept=".pdf" hidden
              onChange={(e) => { setFile(e.target.files?.[0] ?? null); e.target.value = ''; }}
            />
            <button
              type="button"
              className={`filebox w-full ${file ? 'hasfile' : ''}`}
              onClick={() => fileRef.current?.click()}
            >
              {file ? (
                <span className="inline-flex items-center gap-1.5">
                  <FileText className="h-4 w-4" /> {file.name}
                </span>
              ) : 'Clicca per scegliere il PDF dell’ordine…'}
            </button>

            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Mittente (e-mail del cliente)</Label>
              <Input
                placeholder="ordini@cliente.it"
                value={sender}
                onChange={(e) => setSender(e.target.value)}
                list="inboundSenderOptions"
              />
              <datalist id="inboundSenderOptions">
                {senderOptions.map((s) => <option key={s} value={s} />)}
              </datalist>
            </div>
            {match?.found && (
              <p className="text-xs text-emerald-600 dark:text-emerald-400">✓ Template preselezionato: {match.name}</p>
            )}
            {match && !match.found && (
              <p className="text-xs text-amber-600 dark:text-amber-400">
                Nessun template associato a questo mittente: selezionalo a mano.
              </p>
            )}

            <div className="space-y-1.5">
              <Label className="text-xs text-muted-foreground">Template</Label>
              <Select value={tplId || undefined} onValueChange={setTplId}>
                <SelectTrigger>
                  <SelectValue placeholder="— scegli un template —" />
                </SelectTrigger>
                <SelectContent>
                  {templates.map((t) => (
                    <SelectItem key={t.id} value={t.id ?? ''}>{t.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {templates.length === 0 && (
              <div className="flex items-start gap-2 rounded-md border border-amber-300 bg-amber-50 dark:bg-amber-500/10 dark:border-amber-500/30 p-3 text-xs text-amber-800 dark:text-amber-300">
                <Info className="h-4 w-4 shrink-0 mt-0.5" />
                Nessun template definito: creane uno da «Template PDF» prima di importare.
              </div>
            )}

            <div className="flex justify-between mt-2">
              <Button variant="secondary" onClick={() => onOpenChange(false)}>Annulla</Button>
              <Button onClick={extract} disabled={busy || templates.length === 0}>
                {busy ? 'Estrazione…' : 'Estrai dati →'}
              </Button>
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="flex flex-col gap-3">
            <div className="flex flex-wrap gap-2">
              <Badge variant="secondary" className="bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300">
                template: {extraction.tplName}
              </Badge>
              <Badge variant="secondary" className="bg-emerald-100 text-emerald-800 dark:bg-emerald-500/15 dark:text-emerald-300">
                {extraction.n} campi estratti
              </Badge>
              {extraction.low > 0 && (
                <Badge variant="secondary" className="bg-amber-100 text-amber-800 dark:bg-amber-500/15 dark:text-amber-300">
                  {extraction.low} zone vuote o incerte — controlla i valori
                </Badge>
              )}
            </div>

            <div className="grid gap-4 md:grid-cols-[1fr_420px]">
              {/* ── Documento originale ── */}
              <div className="min-w-0 space-y-2">
                <div className="flex items-center gap-2">
                  <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Documento originale
                  </p>
                  {pages.length > 1 && (
                    <div className="ml-auto flex rounded-lg border p-0.5 gap-0.5">
                      {pages.map((p) => (
                        <Button
                          key={p.page_num}
                          variant={p.page_num === curPage ? 'secondary' : 'ghost'}
                          size="sm"
                          className="h-7 px-2 text-xs"
                          onClick={() => setCurPage(p.page_num ?? 0)}
                        >
                          Pag. {(p.page_num ?? 0) + 1}
                        </Button>
                      ))}
                    </div>
                  )}
                </div>
                {currentPage ? (
                  <PdfFieldOverlay page={currentPage} fields={overlayFields} missingTargets={missingTargets} />
                ) : (
                  <div className="flex items-center justify-center rounded-lg border border-dashed p-10 text-center text-xs text-muted-foreground">
                    Anteprima del documento non disponibile.
                  </div>
                )}
              </div>

              {/* ── Campi estratti dall'AI ── */}
              <div className="min-w-0 space-y-3">
                <div className="grid grid-cols-2 gap-2">
                  {DRAFT_KEYS.map(([k, label]) => {
                    const missing = missingTargets.has(k);
                    return (
                      <div key={k} className="space-y-1">
                        <div className="flex min-h-[15px] items-center justify-between gap-1">
                          <Label className="text-[11px] text-muted-foreground uppercase">{label}</Label>
                          {missing && (
                            <span className="whitespace-nowrap text-[10px] font-medium text-destructive">
                              non nel PDF
                            </span>
                          )}
                        </div>
                        <Input
                          className="h-8 text-sm"
                          value={draft[k]}
                          onChange={(e) => setDraft((d) => ({ ...d, [k]: e.target.value }))}
                        />
                      </div>
                    );
                  })}
                </div>
                <div className="space-y-1">
                  <Label className="text-[11px] text-muted-foreground uppercase">Note</Label>
                  <Textarea
                    rows={2}
                    value={draft.notes}
                    onChange={(e) => setDraft((d) => ({ ...d, notes: e.target.value }))}
                  />
                </div>
              </div>
            </div>

            <div className="flex justify-between mt-2">
              <Button variant="secondary" onClick={() => setStep(1)}>← Indietro</Button>
              <div className="flex gap-2">
                <Button variant="secondary" onClick={() => onOpenChange(false)}>Annulla</Button>
                <Button onClick={save} disabled={busy} className="gap-1.5">
                  <Check className="h-4 w-4" /> {busy ? 'Salvataggio…' : 'Crea ordine'}
                </Button>
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
