import { useState } from 'react';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Command, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem } from '@/components/ui/command';
import { Check, ChevronDown } from 'lucide-react';

// Select con ricerca per garage/punti di lavaggio — stile a card (icona
// colorata + nome + indirizzo) invece del semplice testo di un <Select>,
// per allinearsi al mockup "Dettagli trasporto" (ui/TMS Unificato - Standalone.html).
export default function LocationCombobox({ value, onChange, options, placeholder, searchPlaceholder, icon: Icon, iconBg, iconColor, getSubtitle = (o) => o.indirizzo }) {
  const [open, setOpen] = useState(false);
  const selected = options.find(o => o.id === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex w-full items-center gap-2.5 rounded-md border border-dashed px-3 py-2 text-left hover:bg-muted/40"
        >
          <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
            <Icon className="h-3.5 w-3.5" />
          </span>
          <span className="min-w-0 flex-1">
            {selected ? (
              <>
                <span className="block truncate text-sm font-semibold">{selected.nome}</span>
                {getSubtitle(selected) && <span className="block truncate text-xs text-muted-foreground">{getSubtitle(selected)}</span>}
              </>
            ) : (
              <span className="text-sm text-muted-foreground">{placeholder}</span>
            )}
          </span>
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </button>
      </PopoverTrigger>
      {/* z-[1000]: Leaflet's own panes (marker/tooltip/popup) go up to
          z-index 700 — the shadcn default (z-50) leaves the dropdown
          rendering behind the route map whenever one is on the same page. */}
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0 z-[1000]" align="start">
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>Nessun risultato.</CommandEmpty>
            <CommandGroup>
              {options.map(o => (
                <CommandItem
                  key={o.id}
                  value={`${o.nome} ${o.indirizzo || ''}`}
                  onSelect={() => { onChange(o.id); setOpen(false); }}
                >
                  <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full" style={{ background: iconBg, color: iconColor }}>
                    <Icon className="h-3.5 w-3.5" />
                  </span>
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-sm font-semibold">{o.nome}</span>
                    {getSubtitle(o) && <span className="block truncate text-xs text-muted-foreground">{getSubtitle(o)}</span>}
                  </span>
                  {o.id === value && <Check className="h-4 w-4 shrink-0" />}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
