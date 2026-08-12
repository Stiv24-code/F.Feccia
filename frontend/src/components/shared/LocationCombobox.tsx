import type { ComponentType } from 'react';
import SearchableSelect from '@/components/shared/SearchableSelect';

export interface LocationOption {
  id?: string;
  nome?: string;
  indirizzo?: string;
}

export interface LocationComboboxProps<T extends LocationOption> {
  value: string | null | undefined;
  onChange: (id: string) => void;
  options: T[];
  placeholder?: string;
  searchPlaceholder?: string;
  icon: ComponentType<{ className?: string }>;
  iconBg?: string;
  iconColor?: string;
  getSubtitle?: (option: T) => string | undefined;
}

// Select con ricerca per garage/punti di lavaggio — stile a card (icona
// colorata + nome + indirizzo) invece del semplice testo di un <Select>,
// per allinearsi al mockup "Dettagli trasporto" (ui/TMS Unificato - Standalone.html).
// Wrapper specializzato su SearchableSelect: la meccanica popover+ricerca
// vive in un solo posto, qui solo il render "icona + nome + sottotitolo".
export default function LocationCombobox<T extends LocationOption>({ value, onChange, options, placeholder, searchPlaceholder, icon: Icon, iconBg, iconColor, getSubtitle = (o) => o.indirizzo }: LocationComboboxProps<T>) {
  // Nel menu l'icona è colorata (iconBg/iconColor); nel trigger resta grigio
  // neutro anche a selezione fatta — stesso comportamento del componente
  // originale, non legato al colore specifico dell'opzione scelta.
  const renderMenuRow = (o: T) => (
    <>
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full" style={{ background: iconBg, color: iconColor }}>
        <Icon className="h-3.5 w-3.5" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block truncate text-sm font-semibold">{o.nome}</span>
        {getSubtitle(o) && <span className="block truncate text-xs text-muted-foreground">{getSubtitle(o)}</span>}
      </span>
    </>
  );

  const renderTriggerRow = (o: T) => (
    <>
      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <Icon className="h-3.5 w-3.5" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block truncate text-sm font-semibold text-foreground">{o.nome}</span>
        {getSubtitle(o) && <span className="block truncate text-xs text-muted-foreground">{getSubtitle(o)}</span>}
      </span>
    </>
  );

  return (
    <SearchableSelect
      value={value}
      onValueChange={onChange}
      options={options}
      getValue={(o) => o.id || ''}
      getLabel={(o) => o.nome || ''}
      getSearchText={(o) => `${o.nome} ${o.indirizzo || ''}`}
      renderItem={renderMenuRow}
      renderTrigger={renderTriggerRow}
      placeholder={placeholder}
      searchPlaceholder={searchPlaceholder}
      className="h-auto border-dashed py-2 shadow-none hover:bg-muted/40"
    />
  );
}
