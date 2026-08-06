import { useEffect, useRef, useState } from 'react';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Input } from '@/components/ui/input';
import { Loader2 } from 'lucide-react';
import { geocodeSearch } from '@/lib/api';
import { logger } from '@/lib/logger';
import type { DtoGeocodeResultDTO } from '@/api/data-contracts';

const SEARCH_DEBOUNCE_MS = 400;
const MIN_QUERY_LENGTH = 3;

export interface AddressSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  onSelect: (result: DtoGeocodeResultDTO) => void;
  placeholder?: string;
}

// Campo Indirizzo con autocomplete geocoding (ORS/OpenStreetMap): digitando
// appare un elenco di indirizzi candidati (fino a 5 — il match non è mai
// garantito esatto, la copertura dati OSM varia per zona), selezionandone
// uno arrivano anche città/CAP/provincia (quando disponibili) e le
// coordinate per la mappa collegata (vedi MapPicker's flyToSignal). Senza
// virgola tra via/civico e città il parser di ORS (libpostal) spesso non
// segmenta la stringa e degrada a una ricerca full-text che ignora la città
// (verificato: "Via X 4 Comune" trova indirizzi omonimi in città sbagliate,
// "Via X 4, Comune" trova l'indirizzo esatto) — da qui il placeholder col
// formato corretto invece di provare a correggere la query lato nostro.
export function AddressSearchInput({ value, onChange, onSelect, placeholder = 'Es. Via Roma 10, Milano...' }: AddressSearchInputProps) {
  const [results, setResults] = useState<DtoGeocodeResultDTO[]>([]);
  const [searching, setSearching] = useState(false);
  const [open, setOpen] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => () => clearTimeout(debounceRef.current), []);

  const handleChange = (text: string) => {
    onChange(text);
    setOpen(true);
    clearTimeout(debounceRef.current);
    if (text.trim().length < MIN_QUERY_LENGTH) {
      setResults([]);
      setSearching(false);
      return;
    }
    setSearching(true);
    debounceRef.current = setTimeout(() => {
      geocodeSearch(text)
        .then((r: { data: DtoGeocodeResultDTO[] }) => setResults(r.data || []))
        .catch((err: unknown) => { logger.error('Errore ricerca indirizzo:', err); setResults([]); })
        .finally(() => setSearching(false));
    }, SEARCH_DEBOUNCE_MS);
  };

  const handleSelect = (result: DtoGeocodeResultDTO) => {
    onChange(result.indirizzo || '');
    setOpen(false);
    setResults([]);
    onSelect(result);
  };

  return (
    <Popover open={open && (searching || results.length > 0)} onOpenChange={(o) => { if (!o) setOpen(false); }}>
      <PopoverTrigger asChild>
        <div className="relative">
          <Input
            value={value}
            onChange={(e) => handleChange(e.target.value)}
            onFocus={() => setOpen(true)}
            placeholder={placeholder}
            data-testid="address-search-input"
          />
          {searching && <Loader2 className="absolute right-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 animate-spin text-muted-foreground" />}
        </div>
      </PopoverTrigger>
      <PopoverContent
        className="w-[--radix-popover-trigger-width] p-1 z-[1000]"
        align="start"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        {searching ? (
          <div className="flex items-center justify-center gap-2 py-3 text-xs text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" /> Ricerca in corso...
          </div>
        ) : results.length === 0 ? (
          <p className="py-3 text-center text-xs text-muted-foreground">Nessun risultato.</p>
        ) : (
          <div className="flex flex-col gap-0.5 max-h-56 overflow-auto">
            {results.map((r, i) => (
              <button
                key={i}
                type="button"
                onClick={() => handleSelect(r)}
                className="rounded-sm px-2 py-1.5 text-left text-sm hover:bg-muted/60"
              >
                {r.label}
              </button>
            ))}
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
