import { useCallback, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { ArrowLeft, FileText, Pencil, Plus, Trash2, Upload, X } from 'lucide-react';
import { apiClient } from '@/lib/apiClient';
import type {
  DtoPdfRenderPageDTO,
  DtoPdfTemplateFieldDTO,
  DtoPdfTemplateRequest,
  DtoPdfTemplateResponse,
} from '@/api/data-contracts';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import Stage, { type StageZoneNorm } from '@/components/inbound/Stage';
import { TARGETS, getInboundApiError } from '@/components/inbound/constants';

const fmtDate = (iso?: string) => (iso ? new Date(iso).toLocaleDateString('it-IT') : '—');

let uid = 0;
const newLocalId = () => `new-${++uid}`;

// euristica sul testo del blocco per proporre il campo giusto
function guessTarget(text?: string): string {
  const t = (text ?? '').toLowerCase();
  if (/kg|quantit|peso/.test(t)) return 'kg';
  if (/rif|ref|ordine n|order/.test(t)) return 'ref';
  if (/prodotto|merce|product/.test(t)) return 'product';
  if (/carico|load/.test(t)) return 'load_place';
  if (/consegna|scarico|delivery/.test(t)) return 'delivery_place';
  if (/nolo|rate|prezzo|€|eur/.test(t)) return 'rate';
  if (/@/.test(t)) return 'sender_email';
  return 'notes';
}

// Template in modifica nell'editor: come DtoPdfTemplateResponse ma senders
// può momentaneamente essere la stringa grezza del campo di input.
interface EditTemplate {
  id?: string;
  name: string;
  client: string;
  senders: string[] | string;
  fields: DtoPdfTemplateFieldDTO[];
}

interface LoadedPdf {
  file: File;
  pages: DtoPdfRenderPageDTO[];
}

// Editor dei template PDF per cliente (porting OrderMesh): zone disegnabili
// sul PDF renderizzato, collegate ai campi dell'ordine in ingresso; i
// mittenti associati preselezionano il template all'import.
export default function PdfTemplatesPage() {
  const fileRef = useRef<HTMLInputElement>(null);
  const [templates, setTemplates] = useState<DtoPdfTemplateResponse[]>([]);
  const [cur, setCur] = useState<EditTemplate | null>(null);
  const [dirty, setDirty] = useState(false);
  const [pdf, setPdf] = useState<LoadedPdf | null>(null);
  const [curPage, setCurPage] = useState(0);
  const [selField, setSelField] = useState<string | null>(null);
  const [showBlocks, setShowBlocks] = useState(true);
  const [busy, setBusy] = useState<'' | 'render' | 'save' | 'test'>('');
  const [testOut, setTestOut] = useState<Record<string, string> | null>(null);

  const loadTemplates = useCallback(
    () => apiClient.v1PdfTemplatesList()
      .then((res) => setTemplates(res.data ?? []))
      .catch((err) => toast.error(getInboundApiError(err))),
    [],
  );
  useEffect(() => { loadTemplates(); }, [loadTemplates]);

  const confirmDiscard = () =>
    !dirty || window.confirm('Ci sono modifiche non salvate: scartarle?');

  function openTemplate(t: DtoPdfTemplateResponse) {
    if (!confirmDiscard()) return;
    setCur({
      id: t.id,
      name: t.name ?? '',
      client: t.client ?? '',
      senders: [...(t.senders ?? [])],
      fields: (t.fields ?? []).map((f) => ({ ...f })),
    });
    setDirty(false); setPdf(null); setCurPage(0); setSelField(null); setTestOut(null);
  }

  function newTemplate() {
    if (!confirmDiscard()) return;
    setCur({ name: '', client: '', senders: [], fields: [] });
    setDirty(false); setPdf(null); setCurPage(0); setSelField(null); setTestOut(null);
  }

  const patch = (p: Partial<EditTemplate>) => { setCur((c) => (c ? { ...c, ...p } : c)); setDirty(true); };
  const patchField = (id: string, p: Partial<DtoPdfTemplateFieldDTO>) => {
    setCur((c) => (c ? { ...c, fields: c.fields.map((f) => (f.id === id ? { ...f, ...p } : f)) } : c));
    setDirty(true);
  };

  async function uploadSample(file: File) {
    setBusy('render');
    try {
      const res = await apiClient.v1PdfRenderCreate({ file });
      setPdf({ file, pages: res.data.pages ?? [] });
      setCurPage(0);
      toast.success(`PDF caricato: ${res.data.pages?.length ?? 0} pagine, blocchi di testo rilevati`);
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy('');
    }
  }

  function addZone(norm: StageZoneNorm, blockText?: string) {
    const f: DtoPdfTemplateFieldDTO = {
      id: newLocalId(),
      target: guessTarget(blockText),
      label: (blockText || 'campo').slice(0, 40),
      page: curPage,
      x: norm.x, y: norm.y, w: norm.width, h: norm.height,
    };
    setCur((c) => (c ? { ...c, fields: [...c.fields, f] } : c));
    setSelField(f.id ?? null);
    setDirty(true);
  }

  function collectSenders(): string[] {
    if (!cur) return [];
    return typeof cur.senders === 'string'
      ? cur.senders.split(',').map((s) => s.trim()).filter(Boolean)
      : cur.senders;
  }

  function collectRequest(): DtoPdfTemplateRequest {
    return {
      name: (cur?.name ?? '').trim(),
      client: (cur?.client ?? '').trim(),
      senders: collectSenders(),
      // gli id "new-*" sono placeholder della UI: il server ne genera di stabili
      fields: (cur?.fields ?? []).map((f) => ({ ...f, id: f.id?.startsWith('new-') ? '' : f.id })),
    };
  }

  async function save() {
    if (!cur) return;
    const t = collectRequest();
    if (!t.name) { toast.error('Il template deve avere un nome'); return; }
    setBusy('save');
    try {
      const res = cur.id
        ? await apiClient.v1PdfTemplatesUpdate(cur.id, t)
        : await apiClient.v1PdfTemplatesCreate(t);
      const saved = res.data;
      setCur({
        id: saved.id,
        name: saved.name ?? '',
        client: saved.client ?? '',
        senders: [...(saved.senders ?? [])],
        fields: (saved.fields ?? []).map((f) => ({ ...f })),
      });
      setDirty(false);
      await loadTemplates();
      toast.success(`Template salvato: ${saved.name}`);
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy('');
    }
  }

  async function testExtraction() {
    if (!cur) return;
    if (!pdf) { toast.error('Carica prima un PDF di esempio'); return; }
    if (!cur.fields.length) { toast.error('Mappa almeno un campo'); return; }
    setBusy('test');
    try {
      // per il test gli id UI vanno bene (servono solo a correlare le zone)
      const t = { name: cur.name, client: cur.client, senders: collectSenders(), fields: cur.fields };
      const res = await apiClient.v1PdfTestCreate({ file: pdf.file, template: JSON.stringify(t) });
      setTestOut(res.data.values ?? {});
      toast.success('Estrazione di prova completata');
    } catch (err) {
      toast.error(getInboundApiError(err));
    } finally {
      setBusy('');
    }
  }

  async function remove() {
    if (!cur?.id) return;
    if (!window.confirm(`Eliminare il template «${cur.name}»?`)) return;
    try {
      await apiClient.v1PdfTemplatesDelete(cur.id);
      setCur(null); setDirty(false);
      await loadTemplates();
      toast.success('Template eliminato');
    } catch (err) {
      toast.error(getInboundApiError(err));
    }
  }

  // Eliminazione diretta da riga della tabella, senza passare per l'editor.
  async function removeFromList(t: DtoPdfTemplateResponse) {
    if (!t.id) return;
    if (!window.confirm(`Eliminare il template «${t.name}»?`)) return;
    try {
      await apiClient.v1PdfTemplatesDelete(t.id);
      if (cur?.id === t.id) { setCur(null); setDirty(false); }
      await loadTemplates();
      toast.success('Template eliminato');
    } catch (err) {
      toast.error(getInboundApiError(err));
    }
  }

  const page = pdf?.pages?.find((p) => p.page_num === curPage) ?? null;
  const sendersValue = typeof cur?.senders === 'string' ? cur.senders : (cur?.senders ?? []).join(', ');

  // ── Vista elenco (landing della tab Template PDF) ──
  if (!cur) {
    return (
      <div data-testid="pdf-templates-page" className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-muted-foreground">
            Ogni cliente invia i propri ordini PDF con un layout diverso: disegna le zone
            e collegale ai campi dell’ordine. Il template giusto viene scelto dal mittente.
          </p>
          <Button size="sm" className="gap-1.5 text-xs" onClick={newTemplate}>
            <Plus className="h-3.5 w-3.5" /> Nuovo template
          </Button>
        </div>

        <Card className="rounded-xl border shadow-sm">
          <div className="overflow-x-auto">
            <Table className="text-xs md:text-sm">
              <TableHeader>
                <TableRow>
                  <TableHead className="py-2 text-xs">Cliente</TableHead>
                  <TableHead className="py-2 text-xs">Template</TableHead>
                  <TableHead className="py-2 text-xs">Mittente</TableHead>
                  <TableHead className="py-2 text-xs">Blocchi mappati</TableHead>
                  <TableHead className="py-2 text-xs">Aggiornato</TableHead>
                  <TableHead className="py-2 text-xs w-20" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {templates.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} className="py-8 text-center text-muted-foreground">
                      Nessun template ancora.
                    </TableCell>
                  </TableRow>
                )}
                {templates.map((t) => (
                  <TableRow key={t.id} className="cursor-pointer hover:bg-muted/60" onClick={() => openTemplate(t)}>
                    <TableCell className="py-2 font-medium">{t.client || '—'}</TableCell>
                    <TableCell className="py-2">{t.name}</TableCell>
                    <TableCell className="py-2 max-w-[220px] truncate text-muted-foreground">
                      {(t.senders ?? []).join(', ') || '—'}
                    </TableCell>
                    <TableCell className="py-2">
                      <Badge variant="outline">{(t.fields ?? []).length} blocchi</Badge>
                    </TableCell>
                    <TableCell className="py-2 text-muted-foreground">{fmtDate(t.updated_at)}</TableCell>
                    <TableCell className="py-2" onClick={(e) => e.stopPropagation()}>
                      <div className="flex gap-1">
                        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openTemplate(t)}>
                          <Pencil className="h-3 w-3" />
                        </Button>
                        <Button
                          variant="ghost" size="icon" className="h-7 w-7 text-destructive"
                          onClick={() => removeFromList(t)}
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div data-testid="pdf-templates-page" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Ogni cliente invia i propri ordini PDF con un layout diverso: disegna le zone
          e collegale ai campi dell’ordine. Il template giusto viene scelto dal mittente.
        </p>
        <Button
          variant="ghost" size="sm" className="gap-1.5 text-xs"
          onClick={() => { if (confirmDiscard()) { setCur(null); setDirty(false); } }}
        >
          <ArrowLeft className="h-3.5 w-3.5" /> Torna all’elenco
        </Button>
      </div>

      <div className="grid gap-4 items-start md:grid-cols-[240px_1fr_300px]">
        {/* ── Lista template ── */}
        <Card className="p-3">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Template</p>
          <div className="mt-2 flex flex-col gap-1">
            {templates.length === 0 && (
              <p className="text-xs text-muted-foreground">Nessun template ancora.</p>
            )}
            {templates.map((t) => (
              <button
                key={t.id}
                onClick={() => openTemplate(t)}
                className={`flex w-full items-center justify-between gap-2 rounded-lg px-2 py-1.5 text-left transition-colors ${
                  cur?.id === t.id ? 'bg-accent text-accent-foreground' : 'hover:bg-muted'
                }`}
              >
                <span className="min-w-0">
                  <span className="block truncate text-sm">{t.name}</span>
                  <span className="block truncate text-xs text-muted-foreground">
                    {(t.senders ?? []).join(', ') || 'nessun mittente'}
                  </span>
                </span>
                <Badge variant="outline" className="shrink-0">{(t.fields ?? []).length}</Badge>
              </button>
            ))}
            <Button size="sm" className="mt-2 gap-1.5 text-xs" onClick={newTemplate}>
              <Plus className="h-3.5 w-3.5" /> Nuovo template
            </Button>
          </div>
        </Card>

        {/* ── Editor ── */}
        <Card className="p-3">
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <input
              ref={fileRef} type="file" accept=".pdf" hidden
              onChange={(e) => { const f = e.target.files?.[0]; if (f) uploadSample(f); e.target.value = ''; }}
            />
            <Button variant="outline" size="sm" className="gap-1.5 text-xs" onClick={() => fileRef.current?.click()} disabled={busy === 'render'}>
              <Upload className="h-3.5 w-3.5" /> {busy === 'render' ? 'Rendering…' : 'Carica PDF di esempio'}
            </Button>
            {pdf && pdf.pages.length > 1 && (
              <div className="flex rounded-lg border p-0.5 gap-0.5">
                {pdf.pages.map((p) => (
                  <Button
                    key={p.page_num}
                    variant={p.page_num === curPage ? 'secondary' : 'ghost'}
                    size="sm"
                    className="h-7 px-2 text-xs"
                    onClick={() => { setCurPage(p.page_num ?? 0); setSelField(null); }}
                  >
                    Pag. {(p.page_num ?? 0) + 1}
                  </Button>
                ))}
              </div>
            )}
            <div className="ml-auto flex items-center gap-1.5">
              <Checkbox id="show-blocks" checked={showBlocks} onCheckedChange={(v) => setShowBlocks(!!v)} />
              <Label htmlFor="show-blocks" className="text-xs text-muted-foreground cursor-pointer">
                mostra testo rilevato
              </Label>
            </div>
          </div>

          <Stage
            page={page}
            fields={cur.fields.filter((f) => f.page === curPage)}
            selectedId={selField}
            showBlocks={showBlocks}
            onSelect={setSelField}
            onAddZone={addZone}
            onMoveZone={patchField}
            onCommit={() => setDirty(true)}
            onPick={() => fileRef.current?.click()}
          />
        </Card>

        {/* ── Proprietà: dati template + blocchi rilevati, in due card distinte ── */}
        {cur && (
          <div className="flex flex-col gap-3">
          <Card className="p-3">
            <div className="flex flex-col gap-2">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                Dati template
              </p>
              <div className="space-y-1">
                <Label className="text-[11px] text-muted-foreground uppercase">Nome</Label>
                <Input
                  className="h-8 text-sm"
                  placeholder="es. Barilla — ordine trasporto"
                  value={cur.name}
                  onChange={(e) => patch({ name: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <Label className="text-[11px] text-muted-foreground uppercase">Cliente (per la dashboard)</Label>
                <Input
                  className="h-8 text-sm"
                  placeholder="es. BARILLA SPA"
                  value={cur.client}
                  onChange={(e) => patch({ client: e.target.value })}
                />
              </div>
              <div className="space-y-1">
                <Label className="text-[11px] text-muted-foreground uppercase">Mittenti associati</Label>
                <Input
                  className="h-8 text-sm"
                  placeholder="ordini@cliente.it, @cliente.it"
                  value={sendersValue}
                  onChange={(e) => patch({ senders: e.target.value })}
                />
                <p className="text-[11px] text-muted-foreground">
                  Indirizzi separati da virgola. «@dominio.it» abbina l’intero dominio.
                </p>
              </div>
            </div>
          </Card>

          <Card className="p-3">
            <div className="flex flex-col gap-2">
              <p className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                Blocchi rilevati · associazione ai campi ({cur.fields.length})
              </p>
              {cur.fields.length === 0 && (
                <p className="text-xs text-muted-foreground">
                  Nessun campo. Carica un PDF e clicca un blocco di testo o disegna una zona.
                </p>
              )}
              {cur.fields.map((f) => (
                <Card
                  key={f.id}
                  role="button"
                  tabIndex={0}
                  className={`p-2 cursor-pointer ${f.id === selField ? 'ring-2 ring-amber-500' : ''}`}
                  onClick={() => { setSelField(f.id ?? null); if (f.page !== curPage) setCurPage(f.page ?? 0); }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      setSelField(f.id ?? null);
                      if (f.page !== curPage) setCurPage(f.page ?? 0);
                    }
                  }}
                >
                  <div className="flex flex-col gap-1">
                    <Input
                      className="h-7 text-xs"
                      placeholder="etichetta"
                      value={f.label ?? ''}
                      onClick={(e) => e.stopPropagation()}
                      onChange={(e) => f.id && patchField(f.id, { label: e.target.value })}
                    />
                    <div className="flex items-center gap-1">
                      {/* role=presentation: ferma solo la propagazione del clic
                          verso la card; l'interazione vera è la Select dentro */}
                      <div className="flex-1" role="presentation" onClick={(e) => e.stopPropagation()}>
                        <Select value={f.target} onValueChange={(v) => f.id && patchField(f.id, { target: v })}>
                          <SelectTrigger className="h-7 text-xs">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {TARGETS.map(([k, l]) => <SelectItem key={k} value={k}>{l}</SelectItem>)}
                          </SelectContent>
                        </Select>
                      </div>
                      <span className="text-[11px] text-muted-foreground">pag.{(f.page ?? 0) + 1}</span>
                      <Button
                        variant="ghost" size="icon" className="h-6 w-6 text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setCur((c) => (c ? { ...c, fields: c.fields.filter((x) => x.id !== f.id) } : c));
                          if (selField === f.id) setSelField(null);
                          setDirty(true);
                        }}
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                </Card>
              ))}

              <Button size="sm" className="mt-2 text-xs" onClick={save} disabled={busy === 'save'}>
                {busy === 'save' ? 'Salvataggio…' : 'Salva template'}
              </Button>
              <Button variant="outline" size="sm" className="gap-1.5 text-xs" onClick={testExtraction} disabled={busy === 'test'}>
                <FileText className="h-3.5 w-3.5" /> {busy === 'test' ? 'Estrazione…' : 'Prova estrazione sul PDF caricato'}
              </Button>
              {cur.id && (
                <Button
                  variant="outline" size="sm"
                  className="gap-1.5 text-xs text-destructive border-destructive/40 hover:bg-destructive/10"
                  onClick={remove}
                >
                  <Trash2 className="h-3.5 w-3.5" /> Elimina template
                </Button>
              )}

              {testOut && (
                <Card className="bg-muted/50 p-2">
                  <pre className="m-0 max-h-64 overflow-auto whitespace-pre-wrap font-mono text-xs">
                    {TARGETS.filter(([k]) => testOut[k]).map(([k, l]) => `${l}: ${testOut[k]}`).join('\n')
                      || '(nessun valore estratto)'}
                  </pre>
                </Card>
              )}
            </div>
          </Card>
          </div>
        )}
      </div>
    </div>
  );
}
