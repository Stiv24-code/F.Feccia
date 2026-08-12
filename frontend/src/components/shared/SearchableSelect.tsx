import { useState, type ReactNode } from 'react';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Command, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem } from '@/components/ui/command';
import { Check, ChevronsUpDown } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface SearchableSelectProps<T> {
  value?: string | null;
  onValueChange: (value: string) => void;
  options: T[];
  getValue: (item: T) => string;
  getLabel: (item: T) => string;
  /** Testo su cui filtra la ricerca, se diverso dalla label (es. nome + targa). */
  getSearchText?: (item: T) => string;
  /** Render della riga nel menu; default: solo la label. */
  renderItem?: (item: T, selected: boolean) => ReactNode;
  /** Render del valore selezionato nel trigger; default: uguale a renderItem/label. */
  renderTrigger?: (item: T) => ReactNode;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  className?: string;
  contentClassName?: string;
  triggerTestId?: string;
}

// Select con ricerca, drop-in al posto di <Select> di shadcn per le liste
// lunghe (clienti, destinazioni, mezzi, autisti, prodotti...) dove scorrere
// senza filtro è scomodo. Stessa combinazione Popover+Command usata da
// LocationCombobox (che ora si appoggia a questo componente per non
// duplicare la logica) — qui generalizzata su un tipo T qualsiasi invece di
// richiedere lo shape {id, nome, indirizzo}.
export default function SearchableSelect<T>({
  value, onValueChange, options, getValue, getLabel, getSearchText, renderItem, renderTrigger,
  placeholder = 'Seleziona...', searchPlaceholder = 'Cerca...', emptyMessage = 'Nessun risultato.',
  disabled, className, contentClassName, triggerTestId,
}: SearchableSelectProps<T>) {
  const [open, setOpen] = useState(false);
  const selected = options.find((o) => getValue(o) === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          disabled={disabled}
          data-testid={triggerTestId}
          className={cn(
            'flex h-9 w-full items-center justify-between gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
            className,
          )}
        >
          <span className={cn('min-w-0 flex-1 truncate text-left', !selected && 'text-muted-foreground')}>
            {selected ? (renderTrigger ? renderTrigger(selected) : renderItem ? renderItem(selected, false) : getLabel(selected)) : placeholder}
          </span>
          <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      {/* z-[1000]: alcune pagine (AssignOrderForm) mostrano una mappa Leaflet
          sotto — i suoi pane vanno fino a z-700, oltre lo z-50 di default. */}
      <PopoverContent className={cn('w-[--radix-popover-trigger-width] p-0 z-[1000]', contentClassName)} align="start">
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyMessage}</CommandEmpty>
            <CommandGroup>
              {options.map((o) => {
                const v = getValue(o);
                const isSelected = v === value;
                return (
                  <CommandItem
                    key={v}
                    value={getSearchText ? getSearchText(o) : getLabel(o)}
                    onSelect={() => { onValueChange(v); setOpen(false); }}
                  >
                    <div className="flex min-w-0 flex-1 items-center gap-2">
                      {renderItem ? renderItem(o, isSelected) : <span className="truncate">{getLabel(o)}</span>}
                    </div>
                    {isSelected && <Check className="h-4 w-4 shrink-0" />}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
