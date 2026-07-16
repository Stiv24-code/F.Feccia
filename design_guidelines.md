{
  "meta": {
    "product_name": "LoginBusiness",
    "company": "FECCIA F.lli",
    "domain": "Transport Management System (TMS)",
    "ui_language": "it-IT",
    "design_goals": [
      "Operatore-first: velocità, leggibilità, densità controllata",
      "Enterprise/professionale, moderno ma non flashy",
      "Griglie/tabelle scalabili con filtri persistenti",
      "Sidebar scura + contenuto chiaro per ridurre affaticamento"
    ]
  },
  "brand_attributes": [
    "Affidabile",
    "Operativo",
    "Preciso",
    "Sobrio",
    "Performante"
  ],
  "visual_personality": {
    "style_fusion": [
      "Enterprise SaaS (gerarchie nette + moduli ripetibili)",
      "Swiss / International Typographic Style (griglia, allineamenti, densità)",
      "Bento grid molto sobria per KPI (solo dove serve)",
      "Dark sidebar ‘control center’ + light workspace"
    ],
    "do_not": [
      "Niente look consumer (pillole troppo playful, gradient aggressivi)",
      "Niente gradient scuri/saturi, niente aree gradient >20% viewport",
      "Niente layout centrati tipo landing page"
    ]
  },
  "design_tokens": {
    "fonts": {
      "heading": {
        "google_font": "Space Grotesk",
        "tailwind": "font-[\"Space_Grotesk\",system-ui,sans-serif]",
        "usage": "Titoli, KPI numbers, intestazioni moduli"
      },
      "body": {
        "google_font": "IBM Plex Sans",
        "tailwind": "font-[\"IBM_Plex_Sans\",system-ui,sans-serif]",
        "usage": "Testo UI, tabelle, form labels"
      },
      "mono": {
        "google_font": "IBM Plex Mono",
        "tailwind": "font-[\"IBM_Plex_Mono\",ui-monospace,monospace]",
        "usage": "Codici ordine/viaggio, ID fattura, targhe"
      },
      "notes": "Import via Google Fonts in index.html (o @import in CSS) e applicare a body + headings. Evitare troppi pesi: 400/500/600."
    },
    "type_scale": {
      "h1": "text-4xl sm:text-5xl lg:text-6xl",
      "h2": "text-base md:text-lg",
      "body": "text-sm md:text-base",
      "small": "text-xs",
      "table": "text-xs md:text-sm",
      "kpi_number": "text-2xl md:text-3xl"
    },
    "radii": {
      "--radius": "0.5rem",
      "sidebar": "rounded-xl",
      "cards": "rounded-xl",
      "inputs": "rounded-md",
      "chips": "rounded-full"
    },
    "shadows": {
      "card": "shadow-[0_1px_0_rgba(16,24,40,0.06),0_1px_2px_rgba(16,24,40,0.08)]",
      "sidebar": "shadow-[0_8px_24px_rgba(0,0,0,0.28)]",
      "floating": "shadow-[0_18px_40px_rgba(0,0,0,0.20)]"
    },
    "spacing_system": {
      "page_padding": "px-3 sm:px-4 lg:px-6",
      "section_gap": "gap-4 lg:gap-6",
      "card_padding": "p-4 lg:p-5",
      "table_density": {
        "compact": "py-2",
        "standard": "py-3"
      },
      "notes": "Per schermi data-heavy, preferire compact nelle tabelle e standard nei form."
    },
    "color_system": {
      "strategy": "Sidebar dark + workspace light. Stati evidenti ma non fluorescenti. Un solo accento brand (blu petrolio).",
      "css_variables_suggestion": {
        "light": {
          "--background": "210 20% 98%",
          "--foreground": "222 30% 12%",
          "--card": "0 0% 100%",
          "--card-foreground": "222 30% 12%",
          "--muted": "210 18% 96%",
          "--muted-foreground": "215 16% 38%",
          "--border": "214 18% 88%",
          "--input": "214 18% 88%",
          "--ring": "195 92% 28%",
          "--primary": "195 92% 28%",
          "--primary-foreground": "0 0% 100%",
          "--secondary": "210 18% 96%",
          "--secondary-foreground": "222 30% 12%",
          "--accent": "195 40% 92%",
          "--accent-foreground": "222 30% 12%",
          "--destructive": "0 74% 45%",
          "--destructive-foreground": "0 0% 100%"
        },
        "dark": {
          "--background": "222 22% 9%",
          "--foreground": "210 20% 98%",
          "--card": "222 22% 11%",
          "--card-foreground": "210 20% 98%",
          "--muted": "222 18% 15%",
          "--muted-foreground": "215 14% 72%",
          "--border": "222 16% 18%",
          "--input": "222 16% 18%",
          "--ring": "195 92% 48%",
          "--primary": "195 92% 48%",
          "--primary-foreground": "222 22% 9%",
          "--secondary": "222 18% 15%",
          "--secondary-foreground": "210 20% 98%",
          "--accent": "195 35% 18%",
          "--accent-foreground": "210 20% 98%",
          "--destructive": "0 62% 42%",
          "--destructive-foreground": "0 0% 100%"
        }
      },
      "sidebar_palette": {
        "sidebar_bg": "#0B1220",
        "sidebar_bg_2": "#0F1A2E",
        "sidebar_text": "#E7EEF9",
        "sidebar_muted": "#9FB0C7",
        "sidebar_border": "rgba(231,238,249,0.10)",
        "active_item_bg": "rgba(34,211,238,0.10)",
        "active_item_ring": "rgba(34,211,238,0.35)"
      },
      "status_colors": {
        "to_plan": {
          "label": "Da pianificare",
          "semantic": "warning",
          "bg": "#FFF4CC",
          "fg": "#6A4B00",
          "border": "#F1D98A",
          "dot": "#F0B429",
          "tailwind": "bg-[#FFF4CC] text-[#6A4B00] border-[#F1D98A]"
        },
        "assigned": {
          "label": "Assegnato",
          "semantic": "danger",
          "bg": "#FFE2E2",
          "fg": "#7A1F1F",
          "border": "#F3B8B8",
          "dot": "#E24A4A",
          "tailwind": "bg-[#FFE2E2] text-[#7A1F1F] border-[#F3B8B8]"
        },
        "completed_closed": {
          "label": "Completato/Chiuso",
          "semantic": "info",
          "bg": "#FFE9D6",
          "fg": "#7A3B08",
          "border": "#F6C7A0",
          "dot": "#F28B2C",
          "tailwind": "bg-[#FFE9D6] text-[#7A3B08] border-[#F6C7A0]"
        },
        "invoiced": {
          "label": "Fatturato",
          "semantic": "success",
          "bg": "#DFF7E9",
          "fg": "#0F5132",
          "border": "#BFECD2",
          "dot": "#2AA36B",
          "tailwind": "bg-[#DFF7E9] text-[#0F5132] border-[#BFECD2]"
        }
      }
    }
  },
  "layout_and_grids": {
    "app_shell": {
      "pattern": "Left sidebar (dark, collapsible) + topbar (light) + scrollable main content",
      "desktop": {
        "sidebar_width": "w-[280px] (collapsed w-[76px])",
        "topbar_height": "h-14",
        "main_max_width": "max-w-none (data-heavy)",
        "content_padding": "px-3 sm:px-4 lg:px-6 py-4"
      },
      "mobile": {
        "pattern": "Sidebar becomes Sheet/Drawer; topbar shows hamburger + title + search",
        "notes": "Priorità: accesso rapido a Raccolta Ordini, Planner, Fatturazione."
      }
    },
    "dashboard": {
      "kpi_grid": "grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4",
      "secondary_grid": "grid grid-cols-1 xl:grid-cols-[1.6fr_1fr] gap-4",
      "tables": "Preferire una tabella ‘Recenti’ con 6-10 righe e CTA ‘Vedi tutti’"
    },
    "data_pages": {
      "page_header": "Titolo + breadcrumb + azioni (Nuovo, Importa, Esporta) + search",
      "filters": "Row di filtri sticky sotto topbar (opzionale), con reset + salva vista",
      "table_region": "Card contenitore con ScrollArea orizzontale se necessario"
    },
    "planner": {
      "pattern": "Split view: left ‘Ordini non pianificati’ (stack) + right ‘Calendario/Griglia’",
      "grid": "Time columns (giorni) + lane rows (mezzi/viaggi). Drag & drop ordini → lane",
      "density": "Righe alte 48-56px, colonne giorno 180-240px (desktop)"
    }
  },
  "components": {
    "component_path": {
      "core": [
        "/app/frontend/src/components/ui/button.jsx",
        "/app/frontend/src/components/ui/input.jsx",
        "/app/frontend/src/components/ui/label.jsx",
        "/app/frontend/src/components/ui/card.jsx",
        "/app/frontend/src/components/ui/badge.jsx",
        "/app/frontend/src/components/ui/table.jsx",
        "/app/frontend/src/components/ui/tabs.jsx",
        "/app/frontend/src/components/ui/scroll-area.jsx",
        "/app/frontend/src/components/ui/separator.jsx",
        "/app/frontend/src/components/ui/dropdown-menu.jsx",
        "/app/frontend/src/components/ui/dialog.jsx",
        "/app/frontend/src/components/ui/sheet.jsx",
        "/app/frontend/src/components/ui/select.jsx",
        "/app/frontend/src/components/ui/calendar.jsx",
        "/app/frontend/src/components/ui/sonner.jsx",
        "/app/frontend/src/components/ui/tooltip.jsx",
        "/app/frontend/src/components/ui/pagination.jsx",
        "/app/frontend/src/components/ui/resizable.jsx"
      ],
      "forms": [
        "/app/frontend/src/components/ui/form.jsx",
        "/app/frontend/src/components/ui/textarea.jsx",
        "/app/frontend/src/components/ui/checkbox.jsx",
        "/app/frontend/src/components/ui/radio-group.jsx",
        "/app/frontend/src/components/ui/switch.jsx",
        "/app/frontend/src/components/ui/input-otp.jsx"
      ],
      "power_user": [
        "/app/frontend/src/components/ui/command.jsx",
        "/app/frontend/src/components/ui/menubar.jsx",
        "/app/frontend/src/components/ui/context-menu.jsx",
        "/app/frontend/src/components/ui/hover-card.jsx"
      ]
    },
    "component_recipes": {
      "sidebar_nav_item": {
        "description": "Elemento nav con icon + label + badge contatore. Stato attivo con bordo sinistro e bg tenue.",
        "tailwind": "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-[var(--sidebar-muted)] hover:text-[var(--sidebar-text)] hover:bg-white/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300/40",
        "active_tailwind": "bg-[rgba(34,211,238,0.10)] text-[var(--sidebar-text)] ring-1 ring-[rgba(34,211,238,0.25)]",
        "data_testid": "sidebar-nav-item-{route}"
      },
      "status_badge": {
        "base": "inline-flex items-center gap-2 border px-2.5 py-1 text-xs font-medium",
        "dot": "h-2 w-2 rounded-full",
        "data_testid": "order-status-badge",
        "notes": "Usare <Badge> shadcn ma con className custom per rispettare palette stato."
      },
      "dense_table": {
        "description": "Tabelle scorrevoli, header sticky, righe hover e selezione. Supporto colonne fissate se necessario.",
        "tailwind_container": "rounded-xl border bg-card",
        "tailwind_table": "text-xs md:text-sm",
        "row": "hover:bg-muted/60 data-[state=selected]:bg-accent",
        "cell": "py-2",
        "header": "sticky top-0 z-10 bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80",
        "data_testid": "data-table"
      },
      "filter_bar": {
        "description": "Barra filtri compatta con Input search, Select, date range (Calendar in Popover), e pulsanti Salva vista/Reset.",
        "layout": "flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between",
        "data_testid": "filter-bar"
      },
      "kpi_card": {
        "description": "KPI cards sobrie con numero grande, delta, e mini sparkline opzionale.",
        "tailwind": "rounded-xl border bg-card p-4 lg:p-5",
        "data_testid": "kpi-card"
      }
    }
  },
  "micro_interactions_and_motion": {
    "principles": [
      "Motion = feedback, non decorazione",
      "Durate brevi: 120–180ms hover, 180–240ms panel",
      "Easing: ease-out per entrata, ease-in per uscita"
    ],
    "hover_states": {
      "buttons": "hover:brightness-[0.98] active:scale-[0.99]",
      "rows": "hover:bg-muted/60",
      "sidebar": "hover:bg-white/5",
      "chips": "hover:shadow-sm"
    },
    "loading": {
      "pattern": "Skeleton per tabelle e KPI; progress per import/export",
      "component": "/app/frontend/src/components/ui/skeleton.jsx"
    },
    "drag_drop_planner": {
      "notes": "Usare ombra ‘floating’ durante drag + placeholder lane. Snap to grid. Mostrare tooltip con dettaglio ordine.",
      "a11y": "Supportare alternativa keyboard: ‘Assegna’ via Dialog + Select mezzi/viaggi."
    },
    "scroll": {
      "sticky_header": "Header tabella sticky; filter bar sticky solo su desktop",
      "avoid": "No parallax pesante: qui è un tool operazionale."
    }
  },
  "accessibility": {
    "requirements": [
      "WCAG AA contrast per testo e badge",
      "Focus ring visibile (ring + outline) su tutti gli elementi interattivi",
      "Target touch minimo 40px su mobile",
      "Preferenze reduced-motion rispettate"
    ],
    "keyboard": [
      "Command palette (Command) per navigazione rapida moduli",
      "Shortcut suggeriti: / per search, g+d Dashboard, g+p Planner"
    ]
  },
  "data_density_guidelines": {
    "tables": {
      "row_height": "32–40px (compact)",
      "column_alignment": {
        "numbers": "text-right tabular-nums",
        "dates": "whitespace-nowrap",
        "ids": "font-mono"
      },
      "empty_state": {
        "tone": "operativo",
        "copy_it": "Nessun risultato. Modifica i filtri o crea un nuovo record.",
        "cta": "Nuovo"
      }
    },
    "forms": {
      "pattern": "2-column su desktop, single column su mobile",
      "sectioning": "Usare Separator + headings per gruppi (Dati cliente, Dati trasporto, Note)"
    }
  },
  "page_blueprints": {
    "login": {
      "layout": "Split: left brand + trust cues, right form card. Su mobile: stacked.",
      "components": ["Card", "Input", "Label", "Button"],
      "copy_it": {
        "title": "Accedi",
        "subtitle": "Gestione trasporti e pianificazione operativa",
        "primary_cta": "Entra"
      },
      "data_testids": {
        "email": "login-email-input",
        "password": "login-password-input",
        "submit": "login-submit-button"
      }
    },
    "dashboard": {
      "sections": [
        "KPI row",
        "Ordini recenti (tabella)",
        "Andamento (chart) + Eccezioni (lista)"
      ],
      "data_testids": {
        "kpi": "dashboard-kpi",
        "recent_orders": "dashboard-recent-orders"
      }
    },
    "anagrafiche_list": {
      "pattern": "Header + filter bar + dense table + pagination",
      "data_testids": {
        "new": "masterdata-new-button",
        "search": "masterdata-search-input",
        "table": "masterdata-table"
      }
    },
    "planner": {
      "sections": [
        "Toolbar (date range, filtri, salva vista)",
        "Left backlog ordini",
        "Right planning grid"
      ],
      "data_testids": {
        "date_range": "planner-date-range",
        "backlog": "planner-backlog",
        "grid": "planner-grid"
      }
    },
    "fatturazione": {
      "pattern": "Tabs: Proforma | Fatture | Scarti. Tabelle + Dialog dettaglio.",
      "data_testids": {
        "tabs": "invoicing-tabs",
        "export": "invoicing-export-button"
      }
    }
  },
  "charts_and_reporting": {
    "library": {
      "recommended": "recharts",
      "why": "Leggero e veloce per KPI trend e breakdown (per operazioni).",
      "install": "npm i recharts",
      "usage_notes": [
        "Usare palette neutra + accento teal/cyan per serie principali",
        "Tooltip chiaro, no gradient fill aggressivi"
      ]
    },
    "chart_styles": {
      "grid": "stroke: hsl(var(--border))",
      "axis": "tick: hsl(var(--muted-foreground))",
      "series": {
        "primary": "#0EA5A6",
        "secondary": "#1F3A5F"
      }
    }
  },
  "images": {
    "image_urls": [
      {
        "category": "login_background",
        "description": "Immagine astratta/industriale molto sobria per split login (non troppo ‘stocky’).",
        "url": "(preferibile: pattern/texture CSS, evitare foto troppo narrative)"
      }
    ],
    "texture_guidance": {
      "noise_overlay": "Usare noise leggero (opacity 0.04–0.06) su sidebar o hero login; mai su tabelle.",
      "css_snippet": ".noise-overlay{background-image:url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='140' height='140'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='140' height='140' filter='url(%23n)' opacity='.18'/%3E%3C/svg%3E\");opacity:.05;pointer-events:none;}"
    }
  },
  "instructions_to_main_agent": [
    "Aggiornare App.css: rimuovere layout centrato/React default header; non usare .App-header.",
    "Impostare tokens in index.css (HSL) coerenti con sidebar + status colors; aggiungere variabili custom per sidebar.",
    "Creare AppShell React (JS) con Sidebar + Topbar + Main (ScrollArea per contenuti).",
    "Tutte le pagine data-heavy: Header + FilterBar + Table in Card. Persistenza filtri (localStorage) consigliata.",
    "Planner: usare resizable per split panes e implementare drag-drop; prevedere alternativa via Dialog per accessibilità.",
    "Usare Badge per stati ordine con mapping colori: giallo=da pianificare, rosso=assegnato, arancione=chiuso; aggiungere ‘Fatturato’ verde.",
    "Aggiungere data-testid a: sidebar items, search inputs, CTA principali, righe tabella, badge stato, dialog submit.",
    "Usare sonner per toast (successo/salvataggio/errori).",
    "Italiano: labels e microcopy in it-IT (es. ‘Nuovo’, ‘Salva’, ‘Annulla’, ‘Esporta CSV’)."
  ],
  "general_ui_ux_design_guidelines_appendix": "<General UI UX Design Guidelines>\n    - You must **not** apply universal transition. Eg: `transition: all`. This results in breaking transforms. Always add transitions for specific interactive elements like button, input excluding transforms\n    - You must **not** center align the app container, ie do not add `.App { text-align: center; }` in the css file. This disrupts the human natural reading flow of text\n   - NEVER: use AI assistant Emoji characters like`🤖🧠💭💡🔮🎯📚🎭🎬🎪🎉🎊🎁🎀🎂🍰🎈🎨🎰💰💵💳🏦💎🪙💸🤑📊📈📉💹🔢🏆🥇 etc for icons. Always use **FontAwesome cdn** or **lucid-react** library already installed in the package.json\n\n **GRADIENT RESTRICTION RULE**\nNEVER use dark/saturated gradient combos (e.g., purple/pink) on any UI element.  Prohibited gradients: blue-500 to purple 600, purple 500 to pink-500, green-500 to blue-500, red to pink etc\nNEVER use dark gradients for logo, testimonial, footer etc\nNEVER let gradients cover more than 20% of the viewport.\nNEVER apply gradients to text-heavy content or reading areas.\nNEVER use gradients on small UI elements (<100px width).\nNEVER stack multiple gradient layers in the same viewport.\n\n**ENFORCEMENT RULE:**\n    • Id gradient area exceeds 20% of viewport OR affects readability, **THEN** use solid colors\n\n**How and where to use:**\n   • Section backgrounds (not content backgrounds)\n   • Hero section header content. Eg: dark to light to dark color\n   • Decorative overlays and accent elements only\n   • Hero section with 2-3 mild color\n   • Gradients creation can be done for any angle say horizontal, vertical or diagonal\n\n- For AI chat, voice application, **do not use purple color. Use color like light green, ocean blue, peach orange etc**\n\n</Font Guidelines>\n\n- Every interaction needs micro-animations - hover states, transitions, parallax effects, and entrance animations. Static = dead. \n   \n- Use 2-3x more spacing than feels comfortable. Cramped designs look cheap.\n\n- Subtle grain textures, noise overlays, custom cursors, selection states, and loading animations: separates good from extraordinary.\n   \n- Before generating UI, infer the visual style from the problem statement (palette, contrast, mood, motion) and immediately instantiate it by setting global design tokens (primary, secondary/accent, background, foreground, ring, state colors), rather than relying on any library defaults. Don't make the background dark as a default step, always understand problem first and define colors accordingly\n    Eg: - if it implies playful/energetic, choose a colorful scheme\n           - if it implies monochrome/minimal, choose a black–white/neutral scheme\n\n**Component Reuse:**\n\t- Prioritize using pre-existing components from src/components/ui when applicable\n\t- Create new components that match the style and conventions of existing components when needed\n\t- Examine existing components to understand the project's component patterns before creating new ones\n\n**IMPORTANT**: Do not use HTML based component like dropdown, calendar, toast etc. You **MUST** always use `/app/frontend/src/components/ui/ ` only as a primary components as these are modern and stylish component\n\n**Best Practices:**\n\t- Use Shadcn/UI as the primary component library for consistency and accessibility\n\t- Import path: ./components/[component-name]\n\n**Export Conventions:**\n\t- Components MUST use named exports (export const ComponentName = ...)\n\t- Pages MUST use default exports (export default function PageName() {...})\n\n**Toasts:**\n  - Use `sonner` for toasts\"\n  - Sonner component are located in `/app/src/components/ui/sonner.tsx`\n\nUse 2–4 color gradients, subtle textures/noise overlays, or CSS-based noise to avoid flat visuals.\n</General UI UX Design Guidelines>"
}
