# Effetto vetro del tema Glass — come replicarlo in React

Documento per chi porta il tema nell'applicazione. I valori sono estratti dal tema congelato nel
canvas (`ui/design/TMS Unificato.dc.html`), non scritti a memoria: sono quelli che si vedono nella
build `ui/tms-unificato.html`.

---

## 1. Prima cosa: non copiare la tecnica del prototipo

Nel prototipo il tema è un foglio di stile che aggancia gli elementi **dagli stili inline**, con
selettori come:

```css
/* NON replicare: serve solo nel prototipo */
.tms-glass [style*="box-shadow:0 1px 2px rgba(28,37,52,.07)"],
.tms-glass [style*="box-shadow: rgba(28, 37, 52, 0.07)"] { … }
```

Quella forma esiste per un vincolo preciso: nel canvas il markup è generato e non posso aggiungere
classi, quindi l'unico gancio disponibile è la firma dello stile inline già presente. Ha due
conseguenze fastidiose che in React non hai motivo di ereditare: ogni selettore va scritto in
**doppia forma** (il browser normalizza `#fff` in `rgb(255, 255, 255)` e riordina le ombre, quindi
la stringa sorgente non combacia mai), e ogni regola ha bisogno di `!important` per battere lo
stile inline.

Tu hai le classi e i token: usa quelli. Di questo documento ti serve **la ricetta**, cioè quali
superfici esistono, con che opacità e sfocatura, e i tranelli che ho incontrato — non i selettori.

---

## 2. Il concetto: gerarchia per opacità, non "aggiungi blur"

L'errore tipico è mettere `backdrop-filter: blur()` su tutto. Il risultato è una pappa lattiginosa
dove non si capisce cosa contiene cosa. Il tema funziona perché **l'opacità codifica la profondità**:

```
fondale colorato (mesh)              ← l'unica cosa "colorata"
  └─ contenitore di sezione   .36    ← velo tenue: si vede il fondale attraverso
       └─ card interna        .56    ← più chiara: sta sopra
            └─ menù / popup   .82    ← quasi opaco: deve staccarsi
```

Più un elemento è "vicino" all'utente, più è opaco. E **nessuna superficie ha un bordo bianco**:
la separazione la fanno l'opacità e un'ombra larga e morbida. I bordi bianchi erano nella prima
versione e appesantivano tutto.

Regola pratica: se aggiungi una superficie nuova, chiediti a che livello sta e prendi l'opacità
del livello, non un valore a caso.

---

## 3. I token da aggiungere in `src/index.css`

Il tuo file usa già custom properties in `:root`. Queste sono nello stesso stile — le ho lasciate
in `rgba()` invece che nel formato HSL a canali separati perché il vetro ha bisogno del canale
alfa, che con la convenzione `H S% L%` di shadcn non passa.

```css
@layer base {
  :root {
    /* fondale del tema: i colori sono quelli del logo SBG */
    --glass-bg-1: 53 101 175;      /* blu       */
    --glass-bg-2: 143 179 226;     /* blu chiaro*/
    --glass-bg-3: 72 160 56;       /* verde     */
    --glass-bg-4: 40 88 160;       /* blu scuro */

    /* superfici, dal livello più profondo al più vicino */
    --glass-section:  255 255 255 / .36;   /* contenitore di sezione */
    --glass-card:     255 255 255 / .56;   /* card dentro la sezione */
    --glass-bar:      255 255 255 / .12;   /* testata e barre fisse  */
    --glass-sidebar:  242 248 255 / .42;   /* pannello laterale      */
    --glass-menu:     248 251 255 / .82;   /* menù di riga           */
    --glass-picker:   250 250 250 / .82;   /* picker di pianificazione */
    --glass-option:   255 255 255 / .85;   /* voci dentro un picker  */
    --glass-zebra:    255 255 255 / .30;   /* righe alterne tabella  */

    /* sfocature: crescono col livello */
    --blur-section: 20px;
    --blur-bar:     30px;
    --blur-sidebar: 28px;
    --blur-menu:    44px;
    --blur-picker:  52px;

    /* ombre: larghe e tenui, mai nere */
    --shadow-section: 0 18px 44px rgb(30 60 120 / .10);
    --shadow-card:    0 8px 22px rgb(30 60 120 / .08);
    --shadow-float:   0 14px 40px rgb(30 60 120 / .14);
    --shadow-menu:    0 24px 60px rgb(20 45 90 / .30);

    /* bordi: nessun bianco. In chiaro servono solo dove separano davvero */
    --glass-hairline: 96 116 148 / .30;
  }
}
```

Il fondale, che va sull'elemento più esterno dell'app:

```css
.app-glass {
  background:
    radial-gradient(620px 480px at 12%  8%, rgb(var(--glass-bg-1) / .50), transparent 58%),
    radial-gradient(560px 460px at 88% 12%, rgb(var(--glass-bg-2) / .70), transparent 60%),
    radial-gradient(640px 520px at 82% 88%, rgb(var(--glass-bg-3) / .34), transparent 58%),
    radial-gradient(600px 500px at 12% 92%, rgb(var(--glass-bg-4) / .48), transparent 60%),
    radial-gradient(480px 380px at 50% 45%, rgb(255 255 255 / .75), transparent 72%),
    linear-gradient(160deg, #eaf1fa 0%, #d9e6f5 100%);
}
```

I quattro aloni non sono decorazione: danno al blur qualcosa da sfocare. Su un fondale piatto il
vetro non si legge come vetro.

---

## 4. Le superfici, una per una

| Superficie | Sfondo | Blur | Ombra | Bordo |
|---|---|---|---|---|
| Contenitore di sezione | `255 255 255 / .36` | 20px | `--shadow-section` | nessuno |
| Card interna | `255 255 255 / .56` | 14px | `--shadow-card` | nessuno |
| Testata / barre fisse | `255 255 255 / .12` | 30px | nessuna | nessuno |
| Pannello laterale | `242 248 255 / .42` | 28px | `--shadow-float` | nessuno, `border-radius: 18px`, `margin: 12px` |
| Menù di riga | `248 251 255 / .82` | 44px | `--shadow-menu` | nessuno |
| Picker di pianificazione | `250 250 250 / .82` | 52px | `--shadow-menu` | `1px rgb(255 255 255 / .55)` |
| Voci dentro un picker | `255 255 255 / .85` | — | — | `1px` hairline |
| Righe alterne tabella | `255 255 255 / .30` | — | — | nessuno |

Con `saturate()` accanto al blur il fondale sotto resta vivo: `backdrop-filter: blur(20px) saturate(1.6)`.
Sui picker uso invece `saturate(.65)`, e non è un capriccio — vedi il tranello 4.

Le tabelle non hanno linee di separazione tra le righe: la lettura la tengono la zebra al 30% e il
respiro verticale. I separatori **di sezione** invece restano visibili: quando li togli, distingui i
due casi (nel prototipo l'ho fatto agganciando solo le righe cliccabili).

---

## 5. I cinque tranelli

Sono i punti che mi sono costati tempo. Vale la pena leggerli prima di scrivere il CSS.

### 1. `backdrop-filter` crea uno stacking context

Una superficie con `backdrop-filter` diventa un contesto di impilamento: un popup dentro quella
superficie **non può uscirne**, per quanto alto sia il suo `z-index`. Nel prototipo il picker
dell'autista finiva sotto la card successiva, e il tema classico non aveva il problema perché non
ha blur.

La soluzione: quando la card contiene un overlay aperto, alza la card.

```css
.glass-card:has([data-open="true"]) { position: relative; z-index: 20; }
```

In React è più semplice: metti tu la classe quando lo stato dice che il popup è aperto.

### 2. `position: fixed` non raggiunge il viewport se un antenato ha `backdrop-filter`

Stessa causa. Se vuoi una mappa a tutto schermo come fondale e un antenato ha il blur, il `fixed`
si ancora a quell'antenato, non alla finestra. Serve togliere il blur all'antenato, oppure far
diventare quell'antenato stesso il contenitore.

### 3. Su una superficie di vetro non si può appoggiare niente in cima

La testata dell'app è `position: fixed; top: 0` e passa dietro il pannello laterale. Qualunque
fascia aggiunta in cima al documento le finisce sotto — e siccome la testata è trasparente, ne
prende il colore e sembra un bug del tema. Se ti serve una barra di servizio, mettila in fondo.

### 4. Il vetro riflette ciò che ha sotto, e sotto non è uguale in tutte le pagine

Lo stesso picker sopra una mappa e sopra un form bianco sembra di due colori diversi: filtra due
fondali diversi. Se ti serve che un elemento appaia uguale ovunque, non basta dargli gli stessi
valori: devi **desaturare** ciò che filtra (`saturate(.65)`) e alzare l'opacità del velo, così il
fondale pesa meno. È il motivo dei valori "strani" sui picker.

### 5. Il blur non si vede in tutti gli ambienti

Il WebKit headless (quello dei test automatici) dichiara `backdrop-filter` nel computed style ma
**non lo disegna**. Se verifichi il vetro con screenshot in CI, usa Chromium: con WebKit vedi la
pagina senza sfocature e pensi di aver rotto qualcosa.

E una nota di resa: il blur costa. Su liste lunghe evita di metterlo su ogni riga — nel prototipo
le righe e le colonne del kanban hanno un velo **senza** blur, che a occhio non si distingue.

---

## 6. Modalità scura: qui non copiare il prototipo

Nel prototipo la modalità scura è un trucco: un filtro `invert(.94) hue-rotate(180deg)` su tutta
l'app. Comodo per un mockup, sbagliato per l'applicazione — infatti nel tuo `index.css` hai già
notato l'effetto collaterale sulla sidebar, che diventa lilla slavato perché il filtro non la
esclude.

Tu hai i token: **ridefiniscili**, non invertire. La logica da tenere è che in scuro si scambiano i
ruoli — le superfici diventano scure e semi-trasparenti, e i bordi da invisibili diventano
**luminosi**, perché su fondo scuro è la luce a separare:

```css
.dark {
  --glass-section:  20 32 54 / .50;
  --glass-card:     26 40 66 / .62;
  --glass-bar:      14 24 42 / .55;
  --glass-sidebar:  18 30 52 / .58;
  --glass-menu:     22 34 58 / .90;
  --glass-picker:   24 36 60 / .90;
  --glass-option:   30 44 72 / .88;
  --glass-zebra:    255 255 255 / .04;
  --glass-hairline: 255 255 255 / .12;
}
```

⚠️ **Questi ultimi valori sono una proposta di partenza, non misurati**: nel prototipo il dark non
ha valori propri, quindi non posso darti quelli "veri". Vanno calibrati guardandoli, e verificando
i contrasti.

---

## 7. Contrasti: due token vanno corretti rispetto al tema classico

Su fondo di vetro i grigi chiari perdono contrasto. Nel tema li ho scuriti dopo una misurazione sui
pixel di sei schermate:

| Token | Classico | Glass | Perché |
|---|---|---|---|
| `--tx-4` (etichette) | `#76849f` | **`#57657f`** | sotto 4.5:1 sul vetro |
| `--tx-5` (metadati) | `#9aa7bd` | **`#66748e`** | idem |

Restano sotto soglia, per scelta consapevole e da rivedere quando il tema diventa definitivo: le
percentuali sulla mappa (2.97), le voci non attive del pannello laterale (2.8), alcuni verdi e
ambra nei badge (3.7–4.4). Se in applicazione vuoi essere a norma, questi sono i punti da toccare.

Attenzione al metodo: su una superficie di vetro `getComputedStyle` dice l'alfa dichiarato, non il
colore che l'occhio vede. Per misurare davvero i contrasti bisogna **campionare i pixel** di uno
screenshot.

---

## 8. Il font: una sola famiglia

Il tema usa **Outfit** per tutto — titoli, testo, numeri — con `Inter` come ripiego. Nel canvas è
un token unico sul tema glass:

```css
.tms-app.tms-glass { --font: 'Outfit', Inter, sans-serif; }
```

La gerarchia si fa **con i pesi** (400/500/600/700/800), mai alternando famiglie: non c'è un font
da display separato da quello di testo. Sui numeri che stanno in colonna — tariffe, pesi, km, KPI —
va aggiunto `font-variant-numeric: tabular-nums`, altrimenti le cifre ballano fra le righe.

L'import è uno solo:

```html
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700;800&display=swap">
```

Il tema classico (senza `.tms-glass`) resta su `Inter`.

Nota per chi legge `design_guidelines.md`: fino al 10 settembre 2026 quel file indicava Space
Grotesk + IBM Plex Sans, che non sono i font del prototipo. È stato corretto, ma se trovi in giro
riferimenti a quelle due famiglie sono da ignorare — il riferimento è il canvas.

---

## 9. Dove guardare i valori dal vivo

`ui/style-guide.html` è la guida interattiva del tema: contiene l'app in un riquadro e i controlli
per cambiare i colori, il raggio, il font e l'intensità del vetro. Se apri la console e leggi la
funzione `buildGlass()` trovi la ricetta completa parametrizzata sull'intensità — i valori di questo
documento sono quelli a intensità 100%, che è come è congelato il tema.

Per confrontare due varianti conviene misurare, non fidarsi dell'occhio: nel prototipo l'ho fatto
leggendo il `backgroundColor` calcolato per le geometrie e campionando i pixel per i colori reali.

---

*Preparato da Francesco (design) il 3 settembre 2026, sezione sul font aggiunta il 10 settembre. Per domande sui singoli valori, il riferimento
è il blocco `.tms-glass` nel canvas: è la fonte di verità, la style guide e la build ne derivano.*
