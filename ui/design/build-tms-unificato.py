#!/usr/bin/env python3
"""Genera tms-unificato.html (app standalone) da TMS Unificato.dc.html (canvas Claude Design).

Il canvas resta la fonte di verità: dopo ogni re-import da Claude Design
rilanciare questo script per rigenerare l'app.

Trasformazioni:
- rimuove il chrome del canvas (thumbnail bundler, wrapper dv-*, pillole nav demo)
- l'app passa da card fissa 1360x840 a tutto viewport (100vh)
- runtime dc (support.js) e mappa (route-map.html) restano riferimenti locali
"""
import re
import sys
from pathlib import Path

HERE = Path(__file__).parent
SRC = HERE / "TMS Unificato.dc.html"
OUT = HERE / "tms-unificato.html"

src = SRC.read_text(encoding="utf-8")

# blocco <x-dc>...</x-dc> e script di logica
m = re.search(r"<x-dc>(.*)</x-dc>", src, re.S)
if not m:
    sys.exit("x-dc non trovato")
inner = m.group(1)

script = re.search(r'(<script type="text/x-dc" data-dc-script>.*?</script>)', src, re.S)
if not script:
    sys.exit("data-dc-script non trovato")
script = script.group(1)

# 1. via la thumbnail del bundler
inner = re.sub(r'<template id="__bundler_thumbnail".*?</template>\s*', "", inner, flags=re.S)

# 2. helmet invariato; via il wrapper canvas: da <section class="dv-turn"...>
#    fino all'apertura della card inclusa (comprende dv-thd e la riga pillole mNav)
inner = re.sub(
    r'<section class="dv-turn"[^>]*>.*?<div class="dv-card"[^>]*data-screen-label="2a Versione unificata">\s*',
    "",
    inner,
    flags=re.S,
)

# 3. via i 3 </div> di chiusura (dv-card, dv-opt, dv-opts) + </section> in coda
tail = re.search(r"(\s*</div>\s*</div>\s*</div>\s*</section>\s*)$", inner)
if not tail:
    sys.exit("chiusure wrapper canvas non trovate in coda")
inner = inner[: tail.start()] + "\n"

# 4. app a tutto viewport
inner, n = re.subn(r'(class="tms-app[^"]*" style="[^"]*?height:)840px', r"\g<1>100vh", inner)
if n != 1:
    sys.exit(f"attesa 1 sostituzione height tms-app, trovate {n}")

out = f"""<!DOCTYPE html>
<html lang="it">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TMS Unificato · F.lli Feccia</title>
<script src="./support.js"></script>
<meta name="ext-resource-dependency" content="./route-map.html" data-resource-id="routeMap" />
</head>
<body>
<x-dc>{inner}</x-dc>
{script}
</body>
</html>
"""
OUT.write_text(out, encoding="utf-8")
print(f"OK {OUT.name}: {len(out):,} byte")
