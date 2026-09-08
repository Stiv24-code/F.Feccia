import * as React from 'react';
import { Button, type ButtonProps } from '@/components/ui/button';
import { cn } from '@/lib/utils';

// Bottone-icona dentro una riga di tabella. La cornice chip (classe
// `.row-action` in index.css: 26px, radius 8, --chip-bg + --chip-border) è
// quello che lo fa leggere come un bottone invece che come un glifo
// appoggiato sulla riga — nudo, com'era prima, non si capiva fosse cliccabile.
//
// Resta un <Button variant="ghost"> perché l'hover, il focus ring e il
// supporto `asChild` (serve a DropdownMenuTrigger) arrivano da lì: qui si
// aggiunge solo la cornice, e la size di default va rimossa perché h-9/w-9
// vincerebbe sui 26px del chip.
export const RowActionButton = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, ...props }, ref) => (
    <Button
      ref={ref}
      variant="ghost"
      size={null}
      className={cn('row-action shrink-0', className)}
      {...props}
    />
  ),
);
RowActionButton.displayName = 'RowActionButton';
