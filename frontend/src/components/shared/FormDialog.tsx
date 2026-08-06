import * as React from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Loader2 } from 'lucide-react';

export interface FormDialogProps {
  open: boolean;
  onClose: (open: boolean) => void;
  title: React.ReactNode;
  children: React.ReactNode;
  onSubmit?: () => void;
  loading?: boolean;
  submitLabel?: string;
}

export const FormDialog = ({ open, onClose, title, children, onSubmit, loading, submitLabel }: FormDialogProps) => {
  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle style={{ fontFamily: "'Space Grotesk', sans-serif" }}>{title}</DialogTitle>
        </DialogHeader>
        <form onSubmit={(e) => { e.preventDefault(); onSubmit?.(); }} className="space-y-4">
          {children}
          <DialogFooter className="gap-2">
            <Button type="button" variant="outline" onClick={() => onClose(false)} data-testid="form-cancel-button">Annulla</Button>
            <Button type="submit" disabled={loading} data-testid="form-submit-button">
              {loading && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
              {submitLabel || 'Salva'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
