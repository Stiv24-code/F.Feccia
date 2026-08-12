import { useState, useEffect, useCallback, useMemo } from 'react';
import { getInvoices, createInvoice, finalizeInvoice, deleteInvoice, downloadInvoicePdf, getInvoicePdfUrl, getOrders } from '@/lib/api';
import { useGetCustomersQuery } from '@/store/api/appApi';
import { getApiErrorMessage } from '@/lib/apiError';
import type { DtoInvoiceResponse, DtoOrderResponse, DtoInvoiceLineDTO } from '@/api/data-contracts';
import { formatEuro } from '@/lib/format';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Checkbox } from '@/components/ui/checkbox';
import { Skeleton } from '@/components/ui/skeleton';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { toast } from 'sonner';
import { logger } from '@/lib/logger';
import { Plus, FileText, CheckCircle, Trash2, Eye, Loader2 } from 'lucide-react';

type InvoiceTab = 'proforma' | 'definitive';

export default function InvoicesPage() {
  const [invoices, setInvoices] = useState<DtoInvoiceResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<InvoiceTab>('proforma');
  const [newDialogOpen, setNewDialogOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedInvoice, setSelectedInvoice] = useState<DtoInvoiceResponse | null>(null);
  const [saving, setSaving] = useState(false);

  const [closedOrders, setClosedOrders] = useState<DtoOrderResponse[]>([]);
  // Selettore cliente: serve l'elenco completo, non una pagina.
  const { data: customersPage } = useGetCustomersQuery({ limit: 500 }, { skip: !newDialogOpen });
  const customers = customersPage?.items ?? [];
  const [selectedOrderIds, setSelectedOrderIds] = useState<string[]>([]);
  const [selectedClienteId, setSelectedClienteId] = useState('');

  const fetchInvoices = useCallback(() => {
    setLoading(true);
    getInvoices().then((r: { data: DtoInvoiceResponse[] }) => setInvoices(r.data)).catch((err: unknown) => logger.error('Errore caricamento fatture:', err)).finally(() => setLoading(false));
  }, []);
  useEffect(() => { fetchInvoices(); }, [fetchInvoices]);

  const openNew = () => {
    setSelectedOrderIds([]);
    setSelectedClienteId('');
    getOrders({ stato: 'CHIUSO' }).then((r: { data: DtoOrderResponse[] }) => setClosedOrders(r.data));
    setNewDialogOpen(true);
  };

  const filteredClosedOrders = useMemo(() => (selectedClienteId && selectedClienteId !== 'all') ? closedOrders.filter(o => o.cliente_id === selectedClienteId) : closedOrders, [closedOrders, selectedClienteId]);
  const selectedTotal = useMemo(() => closedOrders.filter(o => o.id && selectedOrderIds.includes(o.id)).reduce((s, o) => s + (o.tariffa || 0), 0), [closedOrders, selectedOrderIds]);

  const toggleOrder = (id: string) => {
    setSelectedOrderIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);
  };

  const handleCreateInvoice = async () => {
    if (selectedOrderIds.length === 0 && !selectedClienteId) { toast.error('Selezionare ordini o un cliente'); return; }
    setSaving(true);
    try {
      const selectedOrders = closedOrders.filter(o => o.id && selectedOrderIds.includes(o.id));
      const clienteId = (selectedClienteId && selectedClienteId !== 'all') ? selectedClienteId : (selectedOrders[0]?.cliente_id || '');
      const righe: DtoInvoiceLineDTO[] = selectedOrders.map(o => ({
        ordine_id: o.id,
        descrizione: `${o.destinazione_carico?.nome} - ${o.destinazione_scarico?.nome}`,
        prodotto: o.items?.[0]?.prodotto?.descrizione || '',
        peso: o.items?.[0]?.peso || 0,
        quantita: 1,
        tariffa: o.tariffa,
        totale: o.tariffa,
        iva_codice: 'N8',
      }));
      const totale = righe.reduce((sum, r) => sum + (r.totale || 0), 0);
      await createInvoice({
        cliente_id: clienteId,
        data_fattura: new Date().toISOString().split('T')[0],
        data_scadenza: '',
        condizioni_pagamento: '',
        righe,
        costi_accessori: [],
        totale_imponibile: totale,
        totale_iva: 0,
        totale,
        tipo: 'ordine',
        note: '',
        ordini_ids: selectedOrderIds,
      });
      toast.success('Fattura proforma creata');
      setNewDialogOpen(false);
      fetchInvoices();
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); } finally { setSaving(false); }
  };

  const handleFinalize = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Finalizzare questa fattura? Gli ordini collegati verranno contrassegnati come fatturati.')) return;
    try {
      await finalizeInvoice(id);
      toast.success('Fattura finalizzata');
      fetchInvoices();
      setDetailOpen(false);
    } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const handleDeleteInvoice = async (id?: string) => {
    if (!id) return;
    if (!window.confirm('Eliminare questa fattura proforma?')) return;
    try { await deleteInvoice(id); toast.success('Eliminata'); fetchInvoices(); } catch (e) { toast.error(getApiErrorMessage(e) || 'Errore'); }
  };

  const handleDownloadPdf = async (invoice: DtoInvoiceResponse) => {
    // Issue #35: per fatture DEFINITIVE archiviate su S3 usiamo il
    // presigned URL (download diretto da bucket, no proxy backend).
    // Per le PROFORMA e fallback in caso di errore, generiamo al volo
    // via /invoices/{id}/pdf.
    if (invoice.pdf_s3_key && invoice.id) {
      try {
        const r = await getInvoicePdfUrl(invoice.id);
        if (r.data?.url) {
          window.open(r.data.url, '_blank', 'noopener');
          return;
        }
      } catch (e) {
        logger.warn('presigned URL fallita, fallback a download proxy', e);
      }
    }
    if (!invoice.id) return;
    try {
      const response = await downloadInvoicePdf(invoice.id);
      // Estrai filename dal Content-Disposition; fallback a numero invoice.
      const cd = response.headers?.['content-disposition'] || '';
      const m = cd.match(/filename="?([^";]+)"?/i);
      const filename = m ? m[1] : `fattura_${invoice.numero || invoice.id}.pdf`;
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (e) {
      toast.error(getApiErrorMessage(e) || 'Errore download PDF');
    }
  };

  const proforma = useMemo(() => invoices.filter(i => i.stato === 'PROFORMA'), [invoices]);
  const definitive = useMemo(() => invoices.filter(i => i.stato === 'DEFINITIVA'), [invoices]);

  return (
    <div className="space-y-3" data-testid="invoices-page">
      <div className="flex justify-between items-center">
        <Tabs value={tab} onValueChange={(v) => setTab(v as InvoiceTab)} data-testid="invoicing-tabs">
          <TabsList>
            <TabsTrigger value="proforma">Proforma ({proforma.length})</TabsTrigger>
            <TabsTrigger value="definitive">Definitive ({definitive.length})</TabsTrigger>
          </TabsList>
        </Tabs>
        <Button size="sm" onClick={openNew} className="text-xs gap-1.5" data-testid="invoice-new-button">
          <Plus className="h-3.5 w-3.5" /> Nuova Fattura
        </Button>
      </div>

      <Card className="rounded-xl border shadow-sm">
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                <TableHead className="py-2 text-xs">Numero</TableHead>
                <TableHead className="py-2 text-xs">Cliente</TableHead>
                <TableHead className="py-2 text-xs">Data</TableHead>
                <TableHead className="py-2 text-xs text-right">Totale</TableHead>
                <TableHead className="py-2 text-xs">Stato</TableHead>
                <TableHead className="py-2 text-xs">Azioni</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={`skel-row-${i}`}>{Array.from({ length: 6 }).map((_, j) => <TableCell key={`skel-col-${j}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>)}</TableRow>
              )) : (tab === 'proforma' ? proforma : definitive).length === 0 ? (
                <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Nessuna fattura</TableCell></TableRow>
              ) : (tab === 'proforma' ? proforma : definitive).map(inv => (
                <TableRow key={inv.id} className="hover:bg-muted/60">
                  <TableCell className="py-2 font-mono font-medium">{inv.numero}</TableCell>
                  <TableCell className="py-2">{inv.cliente?.ragione_sociale}</TableCell>
                  <TableCell className="py-2 whitespace-nowrap">{inv.data_fattura}</TableCell>
                  <TableCell className="py-2 text-right tabular-nums font-medium">€ {formatEuro(inv.totale || 0)}</TableCell>
                  <TableCell className="py-2"><StatusBadge stato={inv.stato} /></TableCell>
                  <TableCell className="py-2">
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => { setSelectedInvoice(inv); setDetailOpen(true); }} aria-label="Apri dettaglio fattura"><Eye className="h-3 w-3" /></Button>
                      <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleDownloadPdf(inv)} title="Scarica PDF" aria-label="Scarica PDF fattura"><FileText className="h-3 w-3" /></Button>
                      {inv.stato === 'PROFORMA' && (
                        <>
                          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleFinalize(inv.id)} title="Finalizza" aria-label="Finalizza fattura"><CheckCircle className="h-3 w-3" /></Button>
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={() => handleDeleteInvoice(inv.id)} aria-label="Elimina fattura"><Trash2 className="h-3 w-3" /></Button>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* New Invoice Dialog */}
      <Dialog open={newDialogOpen} onOpenChange={setNewDialogOpen}>
        <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Nuova Fattura Proforma</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <Label>Filtra per Cliente</Label>
              <Select value={selectedClienteId} onValueChange={setSelectedClienteId}>
                <SelectTrigger><SelectValue placeholder="Tutti i clienti" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Tutti</SelectItem>
                  {customers.map(c => <SelectItem key={c.id} value={c.id || ''}>{c.ragione_sociale}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label className="mb-2 block">Ordini chiusi da fatturare ({filteredClosedOrders.length})</Label>
              <div className="border rounded-lg max-h-48 overflow-y-auto">
                {filteredClosedOrders.length === 0 ? (
                  <p className="text-sm text-muted-foreground p-3">Nessun ordine chiuso disponibile</p>
                ) : filteredClosedOrders.map(o => (
                  <label key={o.id} className="flex items-center gap-2 px-3 py-2 hover:bg-muted/50 cursor-pointer text-sm">
                    <Checkbox checked={!!o.id && selectedOrderIds.includes(o.id)} onCheckedChange={() => o.id && toggleOrder(o.id)} />
                    <span className="font-mono text-xs">{o.progressivo}</span>
                    <span className="truncate flex-1">{o.destinazione_carico?.nome} → {o.destinazione_scarico?.nome}</span>
                    <span className="tabular-nums">€ {formatEuro(o.tariffa || 0)}</span>
                  </label>
                ))}
              </div>
              {selectedOrderIds.length > 0 && (
                <p className="text-sm mt-2 font-medium">Totale: € {formatEuro(selectedTotal)}</p>
              )}
            </div>
          </div>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setNewDialogOpen(false)}>Annulla</Button>
            <Button onClick={handleCreateInvoice} disabled={saving} data-testid="invoice-create-button">
              {saving && <Loader2 className="h-4 w-4 animate-spin mr-2" />} Crea Proforma
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Detail Dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Fattura {selectedInvoice?.numero}</DialogTitle></DialogHeader>
          {selectedInvoice && (
            <div className="space-y-3 text-sm">
              <div className="flex justify-between"><span className="text-muted-foreground">Cliente:</span><span>{selectedInvoice.cliente?.ragione_sociale}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Data:</span><span>{selectedInvoice.data_fattura}</span></div>
              <div className="flex justify-between"><span className="text-muted-foreground">Stato:</span><StatusBadge stato={selectedInvoice.stato} /></div>
              {selectedInvoice.righe && selectedInvoice.righe.length > 0 && (
                <div>
                  <Label className="mb-2 block">Righe fattura:</Label>
                  <div className="border rounded-lg divide-y">
                    {selectedInvoice.righe.map((r, i) => (
                      <div key={r.ordine_id || `riga-${i}`} className="px-3 py-2 flex justify-between">
                        <span className="truncate">{r.descrizione}</span>
                        <span className="tabular-nums font-medium">€ {formatEuro(r.totale || 0)}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              <div className="flex justify-between font-medium text-base pt-2 border-t">
                <span>Totale:</span>
                <span>€ {formatEuro(selectedInvoice.totale || 0)}</span>
              </div>
              {selectedInvoice.stato === 'PROFORMA' && (
                <Button className="w-full mt-2" onClick={() => handleFinalize(selectedInvoice.id)} data-testid="invoice-finalize-button">
                  <CheckCircle className="h-4 w-4 mr-2" /> Finalizza Fattura
                </Button>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
