import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import AssignOrderForm from './AssignOrderForm';

// Modale "Assegna Ordine" — riutilizzata da PlannerPage (calendario/lista).
// Il corpo del form (mezzi/autisti/vettori, disponibilita', punto di
// partenza/lavaggio) vive in AssignOrderForm, condiviso con la vista inline
// di OrderDetailPage.
export default function AssignOrderDialog({ open, onOpenChange, order, onAssigned }) {
  const handleAssigned = () => {
    onOpenChange(false);
    if (onAssigned) onAssigned();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>Assegna Ordine {order?.progressivo}</DialogTitle>
        </DialogHeader>
        <div className="p-3 rounded-lg bg-muted/50 text-sm">
          <p><strong>{order?.destinazione_carico?.nome}</strong> → <strong>{order?.destinazione_scarico?.nome}</strong></p>
          <p className="text-muted-foreground">{order?.cliente?.ragione_sociale} • {order?.data_ritiro}</p>
        </div>
        {order && (
          <AssignOrderForm order={order} onAssigned={handleAssigned} onCancel={() => onOpenChange(false)} />
        )}
      </DialogContent>
    </Dialog>
  );
}
