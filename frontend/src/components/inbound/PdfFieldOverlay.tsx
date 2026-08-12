import type { DtoPdfRenderPageDTO, DtoPdfTemplateFieldDTO } from '@/api/data-contracts';
import { TARGET_LABEL } from './constants';

interface PdfFieldOverlayProps {
  page: DtoPdfRenderPageDTO | null;
  fields: DtoPdfTemplateFieldDTO[];
  missingTargets: Set<string>;
}

const pct = (v?: number) => `${(v ?? 0) * 100}%`;

// Versione read-only dello Stage dell'editor template: stessa immagine di
// pagina + stesso positioning normalizzato 0..1, ma senza drag/resize — qui
// serve solo a mostrare dove l'AI ha letto ogni campo nel documento originale
// durante la revisione dell'ordine importato da PDF.
export default function PdfFieldOverlay({ page, fields, missingTargets }: PdfFieldOverlayProps) {
  if (!page) {
    return (
      <div className="stage">
        <p className="text-center text-sm text-muted-foreground px-5 py-16">
          Rendering del documento non disponibile.
        </p>
      </div>
    );
  }

  return (
    <div className="stage">
      <img
        src={`data:image/png;base64,${page.image_b64}`}
        alt={`pagina ${(page.page_num ?? 0) + 1}`}
      />
      {fields.map((f) => {
        const missing = !!f.target && missingTargets.has(f.target);
        return (
          <div
            key={f.id ?? f.target}
            className={`stage-zone readonly ${missing ? 'missing' : ''}`}
            style={{ left: pct(f.x), top: pct(f.y), width: pct(f.w), height: pct(f.h) }}
          >
            <span className="tag">{TARGET_LABEL[f.target ?? ''] ?? f.target}</span>
          </div>
        );
      })}
    </div>
  );
}
