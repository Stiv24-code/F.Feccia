import { useRef, useState, type PointerEvent as ReactPointerEvent, type KeyboardEvent as ReactKeyboardEvent } from 'react';
import type { DtoPdfRenderPageDTO, DtoPdfTemplateFieldDTO } from '@/api/data-contracts';
import { TARGET_LABEL } from './constants';

export interface StageZoneNorm {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface StageProps {
  page: DtoPdfRenderPageDTO | null;
  fields: DtoPdfTemplateFieldDTO[];
  selectedId: string | null;
  showBlocks: boolean;
  onSelect: (id: string) => void;
  onAddZone: (norm: StageZoneNorm, blockText?: string) => void;
  onMoveZone: (id: string, patch: Partial<DtoPdfTemplateFieldDTO>) => void;
  onCommit: () => void;
}

interface DragState {
  id: string;
  resize: boolean;
  sx: number;
  sy: number;
  ox: number;
  oy: number;
  ow: number;
  oh: number;
  moved: boolean;
}

interface RubberState {
  x0: number;
  y0: number;
  x1: number;
  y1: number;
}

// Stage: immagine della pagina PDF con sopra
//  - i blocchi di testo rilevati (clic → nuovo campo su quella zona)
//  - le zone mappate (trascinabili, ridimensionabili dall'angolo)
//  - il "rubber band" per disegnare una zona nuova su area vuota.
// Tutte le coordinate dei campi sono normalizzate 0..1. Porting 1:1 dello
// Stage di OrderMesh; gli stili .stage-* vivono in App.css.
export default function Stage({
  page, fields, selectedId, showBlocks,
  onSelect, onAddZone, onMoveZone, onCommit,
}: StageProps) {
  const imgRef = useRef<HTMLImageElement>(null);
  const dragRef = useRef<DragState | null>(null);
  const [rubber, setRubber] = useState<RubberState | null>(null);

  const rect = () => imgRef.current?.getBoundingClientRect();

  function startDrag(ev: ReactPointerEvent, field: DtoPdfTemplateFieldDTO, resize: boolean) {
    ev.preventDefault();
    ev.stopPropagation();
    if (!field.id) return;
    onSelect(field.id);
    dragRef.current = {
      id: field.id, resize,
      sx: ev.clientX, sy: ev.clientY,
      ox: field.x ?? 0, oy: field.y ?? 0, ow: field.w ?? 0, oh: field.h ?? 0,
      moved: false,
    };
    window.addEventListener('pointermove', moveDrag);
    window.addEventListener('pointerup', endDrag);
  }

  function moveDrag(ev: PointerEvent) {
    const d = dragRef.current;
    const r = rect();
    if (!d || !r) return;
    d.moved = true;
    const dx = (ev.clientX - d.sx) / r.width;
    const dy = (ev.clientY - d.sy) / r.height;
    if (d.resize) {
      onMoveZone(d.id, {
        w: Math.max(0.005, Math.min(1, d.ow + dx)),
        h: Math.max(0.005, Math.min(1, d.oh + dy)),
      });
    } else {
      onMoveZone(d.id, {
        x: Math.max(0, Math.min(1 - d.ow, d.ox + dx)),
        y: Math.max(0, Math.min(1 - d.oh, d.oy + dy)),
      });
    }
  }

  function endDrag() {
    window.removeEventListener('pointermove', moveDrag);
    window.removeEventListener('pointerup', endDrag);
    if (dragRef.current?.moved) onCommit();
    dragRef.current = null;
  }

  function startRubber(ev: ReactPointerEvent) {
    const r = rect();
    if (!r) return;
    const x0 = ev.clientX - r.left;
    const y0 = ev.clientY - r.top;
    const state: RubberState = { x0, y0, x1: x0, y1: y0 };
    setRubber(state);

    const move = (e: PointerEvent) => {
      state.x1 = Math.max(0, Math.min(r.width, e.clientX - r.left));
      state.y1 = Math.max(0, Math.min(r.height, e.clientY - r.top));
      setRubber({ ...state });
    };
    const up = () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', up);
      setRubber(null);
      const w = Math.abs(state.x1 - state.x0) / r.width;
      const h = Math.abs(state.y1 - state.y0) / r.height;
      if (w > 0.004 && h > 0.004) {
        onAddZone({
          x: Math.min(state.x0, state.x1) / r.width,
          y: Math.min(state.y0, state.y1) / r.height,
          width: w, height: h,
        });
      }
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
  }

  if (!page) {
    return (
      <div className="stage">
        <p className="text-center text-sm text-muted-foreground px-5 py-16">
          Carica un PDF di esempio del cliente per mappare i campi.
          <br />
          <span className="text-xs">
            Clic su un blocco di testo rilevato per aggiungerlo come campo,
            oppure trascina per disegnare una zona.
          </span>
        </p>
      </div>
    );
  }

  const pct = (v?: number) => `${(v ?? 0) * 100}%`;

  return (
    <div
      className="stage"
      onPointerDown={(ev) => {
        const target = ev.target as HTMLElement;
        if (target.closest('.stage-zone') || target.closest('.stage-blk')) return;
        startRubber(ev);
      }}
    >
      <img
        ref={imgRef}
        src={`data:image/png;base64,${page.image_b64}`}
        alt={`pagina ${(page.page_num ?? 0) + 1}`}
      />

      {showBlocks && (page.blocks ?? []).map((b, i) => {
        const n = b.bounds_norm ?? {};
        const add = () =>
          onAddZone({ x: n.x ?? 0, y: n.y ?? 0, width: n.width ?? 0, height: n.height ?? 0 }, b.text);
        return (
          <div
            key={i}
            className="stage-blk"
            title={b.text}
            role="button"
            tabIndex={0}
            style={{ left: pct(n.x), top: pct(n.y), width: pct(n.width), height: pct(n.height) }}
            onClick={(ev) => { ev.stopPropagation(); add(); }}
            onKeyDown={(ev: ReactKeyboardEvent) => {
              if (ev.key === 'Enter' || ev.key === ' ') { ev.preventDefault(); add(); }
            }}
          />
        );
      })}

      {fields.map((f) => (
        <div
          key={f.id}
          className={`stage-zone ${f.id === selectedId ? 'sel' : ''}`}
          style={{ left: pct(f.x), top: pct(f.y), width: pct(f.w), height: pct(f.h) }}
          onPointerDown={(ev) => startDrag(ev, f, false)}
        >
          <span className="tag">{TARGET_LABEL[f.target ?? ''] ?? f.target}</span>
          <span className="rsz" onPointerDown={(ev) => startDrag(ev, f, true)} />
        </div>
      ))}

      {rubber && (
        <div
          className="stage-rubber"
          style={{
            left: Math.min(rubber.x0, rubber.x1),
            top: Math.min(rubber.y0, rubber.y1),
            width: Math.abs(rubber.x1 - rubber.x0),
            height: Math.abs(rubber.y1 - rubber.y0),
          }}
        />
      )}
    </div>
  );
}
