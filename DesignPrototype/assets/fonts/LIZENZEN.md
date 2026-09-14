# Schriftlizenzen

Alle drei Familien liegen als Datei im Projekt und werden per `@font-face`
aus `lifelog.css` geladen. Nichts wird von einem fremden Server nachgeladen.
Der vollständige Lizenztext steht mit den Copyright-Zeilen aller Familien in
`OFL.txt` daneben und wird zusammen mit den Schriften ausgeliefert.

| Familie | Rolle | Lizenz | Stand der Dateien |
|---|---|---|---|
| **Manrope** 400/500/600/700 | Display — Titel, Kennzahlen, Abschnittslabels | SIL OFL 1.1 | Subset, siehe unten |
| **Source Sans 3** 400/600 | Text und Oberfläche, inkl. Datum und KW | SIL OFL 1.1 | unverändertes Original |
| **iA Writer Mono S** 400 | nur Capture-Zeile und Code | SIL OFL 1.1 | unverändertes Original |

Die OFL verlangt dreierlei: den Lizenztext mitliefern, die Schriften nicht
einzeln verkaufen, und reservierte Namen bei Änderungen nicht weiterverwenden.
Punkt eins erfüllt `OFL.txt`. Punkt zwei ist nicht berührt — die Schriften sind
Teil der Anwendung und werden nicht als solche verkauft. Punkt drei ist der
einzige, bei dem der Stand der einzelnen Dateien zählt.

## Manrope

Die vier Dateien sind gegenüber der Ausgabe von Google Fonts auf
OpenType-Feature-Ebene reduziert (8 statt 14 Features, 737 statt 740 Glyphen).
Nach OFL-FAQ 2.6 gilt das als Modified Version. Zulässig ist es trotzdem:
Manrope deklariert keinen Reserved Font Name, weder in `OFL.txt` noch in den
Metadaten der Dateien. Die Namensbeschränkung greift damit nicht. Die Schrift
muss unter der OFL bleiben, was sie hier tut.

Die URL in Manropes Copyright-Zeile, `github.com/sharanda/manrope`, ist nicht
mehr erreichbar. Die Zeile bleibt trotzdem wörtlich stehen — sie ist Teil des
Hinweises, den die Lizenz zu erhalten verlangt. Laut `METADATA.pb` bei Google
Fonts liegt die Schrift heute unter `github.com/aaronbell/manrope`.

## Source Sans 3

`SourceSans3-400.woff2` und `-600.woff2` stammen unverändert aus Adobes
Release 3.052R, Archiv `WOFF2-source-sans-3.052R.zip`, Ordner `WOFF2/TTF`.
Vollständige Schrift mit 2478 Glyphen und 50 OpenType-Features.

Vorher lagen hier subsettete TTF-Dateien mit 1801 Glyphen und 7 Features. Da
Adobe den Namen „Source" als Reserved Font Name führt, hätte eine so gekürzte
Fassung nicht weiter so heißen dürfen. Adobes eigenes WOFF2 löst das und ist
mit 218 KB für beide Schnitte zugleich kleiner als die 469 KB davor.

Adobes Copyright-Zeile enthält zusätzlich einen Markenhinweis. Der ist kein
Bestandteil der OFL, sondern Markenrecht, und steht in `OFL.txt` mit drin.

## iA Writer Mono S

Byte-identisch mit `iA Writer Mono/Static/iAWriterMonoS-Regular.ttf` aus
`github.com/iaolo/iA-Fonts`. Die Schrift leitet sich von IBM Plex ab; beide
Copyright-Zeilen stehen deshalb in `OFL.txt`, so wie iA sie in der eigenen
`LICENSE.md` führt.

Zwei Eigenheiten, die beim Nachprüfen auffallen und keine sind: Die Datei
enthält keine Lizenzfelder in den Metadaten und trägt im Copyright-Feld den
Vermerk „All rights reserved". Maßgeblich ist die `LICENSE.md` im selben
Ordner des Repositorys, und die nennt die OFL 1.1. Ebenfalls uneinheitlich ist
der Rechtsträger: Die Datei nennt „Information Architects GmbH", die
`LICENSE.md` nennt „Information Architects Inc." Übernommen ist die Angabe aus
der `LICENSE.md`.

## Nachprüfen

```sh
# Lizenztext gegen SIL, ohne die Copyright-Zeilen darüber
body() { sed -n '/^This Font Software is licensed under/,$p' "$1"; }
curl -s https://openfontlicense.org/documents/OFL.txt > /tmp/ofl
diff <(body /tmp/ofl) <(body OFL.txt)

# iA Writer Mono S gegen das Original
curl -sL "https://raw.githubusercontent.com/iaolo/iA-Fonts/master/iA%20Writer%20Mono/Static/iAWriterMonoS-Regular.ttf" \
  | shasum -a 256
# 929605302a57250e712908cb5f6e1ce80c7d0accd5fd2555345f29a5e8d4e30b
```

Die ausgelieferten Prüfsummen:

```
929605302a57250e712908cb5f6e1ce80c7d0accd5fd2555345f29a5e8d4e30b  iAWriterMonoS-Regular.ttf
a13d9b41b0a471ce58f0e46d377fa3cc76615e4632c3b15eb397caadf7a13f0a  Manrope-400.ttf
f433e5d333be1128d8ec8a28b8aadd0314d10e14e4105406076666bc830fc15d  Manrope-500.ttf
6bad1a774228464cc88b8b0271555b266f7a3c64a7bedf15e165fdab4f6ac0ce  Manrope-600.ttf
b9584e099e0e7f1e1e914c3037aabb9010c50346a3ce8df5eeeeecc0563044d8  Manrope-700.ttf
53492fb3a0def77354f166a55d09b63a10855e91c206c7620a81cf56e97f8ec3  SourceSans3-400.woff2
47b9b661b9f395fe7f0d0e119637fba5c8dad97bde3df60066fd24229c0792f4  SourceSans3-600.woff2
```

Frühere Stände nutzten URW Gothic (AGPL) und iA Writer Quattro. Beide sind
nicht mehr eingebunden und wurden aus dem Projekt entfernt.
