# lifelog — Design-Plan

## BLUF

**Dein Problem ist keine Design-Frage, sondern eine fehlende These.** Der Screenshot zeigt eine Todo-App mit zwei angebauten leeren Kisten — deshalb fühlt sie sich falsch an, nicht wegen der Farben.

**Die These, die ich empfehle:** *lifelog ist ein Tagebuch, das man ausfüllt, indem man lebt — nicht eine Datenbank, die man pflegt.*

Daraus folgen drei Entscheidungen, die alles andere festlegen:

1. **Der Anker ist „Heute", nicht „Todos".** Ein Screen beantwortet drei Fragen: *Was ist jetzt dran? Was ist heute noch los? Wie lief gestern?* Alles andere ist ein eigener Screen.
2. **Eine Eingabe für alles.** Die Quick-Add-Zeile, die du schon hast (`Wäsche aufhängen 15m #haushalt @05.09.2026`), wird global (Cmd/Ctrl+K) und der *einzige* Weg, etwas anzulegen — Aufgabe, Notiz, Habit-Tick, Termin. Das ist dein Alleinstellungsmerkmal gegenüber Notion und Todoist. Keine Formulare, keine „+ Neu"-Buttons pro Modul.
3. **Der Payoff-Screen ist „Rückblick", nicht „Habits".** Du hast geschrieben: *„Habits mit denen die gut liefen und schlecht liefen"* — das ist keine Liste, das ist eine Auswertung. Tracken ohne Auszahlung ist Hausarbeit. Dieser Screen ist der Grund, warum du die App morgen wieder öffnest.

**Datenmodell in einem Satz:** Alles ist ein Eintrag mit `typ` (task | note | habit | event), optional `projekt`, `tags`, `zeit`. Deshalb funktioniert die eine Eingabezeile, und deshalb sind alle Ansichten nur Filter über denselben Stream.

---

## Was am aktuellen Screen konkret schiefgeht

Damit die Redesign-Entscheidungen nachvollziehbar sind:

| # | Befund | Ursache |
|---|---|---|
| 1 | „Tagesplan — später." / „Freitext-Notizen — später." stehen als UI im Produkt | Platzhalter wurden ausgeliefert statt Sektionen wegzulassen |
| 2 | ~70 % Totraum; Karten auf volle Höhe gestreckt | Layout folgt dem Grid, nicht dem Inhalt |
| 3 | TODOS, NOTIZEN, KALENDER sind gleich laut | Keine Hierarchie — drei Module, kein Hauptmotiv |
| 4 | Titel lautet „Todos" | Verrät die Denkweise: Todo-App mit Anbau |
| 5 | `Fr 4.9.` `2h` `#lifelog` sind so präsent wie der Aufgabentext | Metadaten konkurrieren mit Inhalt statt ihn zu stützen |
| 6 | Icon-Rail ohne Label, aktiver Zustand = dünner Strich | Navigation ohne Orientierung |
| 7 | Keine Empty States, keine sichtbaren Hover-/Fokus-Zustände | Zustände wurden nie entworfen |

**Was gut ist und bleibt:** dunkle, ruhige Grundfläche · Mono für Metadaten · Tag-Chips · die Quick-Add-Syntax · die schmale Icon-Rail als Prinzip.

---

## Verteilung: Was gehört auf „Heute", was wird ausgelagert

Das ist deine eigentliche Frage. Regel: **Auf „Heute" steht nur, was sich heute ändert. Alles Nachschlagbare wandert raus.**

### Auf „Heute" (der Anker-Screen)

- **Jetzt-Block** — 1 bis 3 Aufgaben, groß und ruhig. Nicht 40. Der Rest liegt eine Ebene tiefer.
- **Tagesachse** — die heutigen Termine als vertikale Zeitleiste, mit einer „jetzt"-Markierung. Ersetzt die leere Kalenderkiste.
- **Habit-Streifen** — die heutigen Habits als 5–7 abhakbare Punkte in einer Zeile. Ein Klick, kein Dialog.
- **Tageslog** — freies Markdown-Feld für den Tag. Wird beim Tageswechsel automatisch zur Notiz `2026-09-06`.
- **Capture-Zeile** — unten fixiert, immer erreichbar.

### Eigene Screens (ausgelagert)

| Screen | Beantwortet | Kerninhalt |
|---|---|---|
| **Projekte** | „Woran arbeite ich gerade?" | Liste → Detail mit Markdown-Notizen, zugehörigen Aufgaben, aufgelaufener Zeit |
| **Notizen** | „Was hatte ich dazu aufgeschrieben?" | Markdown-Bibliothek, Volltextsuche, Tag-Filter, `[[Verlinkung]]` |
| **Habits** | „Wie ist die Historie?" | Heatmap pro Habit, Streak, Verlauf über Wochen |
| **Kalender** | „Wie sieht die Woche aus?" | Woche/Monat, Termine + terminierte Aufgaben in einer Ansicht |
| **Rückblick** | „Wie lief die Woche wirklich?" | Was lief gut / was lief schlecht, abgeschlossene Aufgaben, Habit-Quote, Zeit pro Projekt |
| **Einstellungen** | Konfiguration | Kalender-Anbindung, Habits definieren, Tags verwalten |

### Bewusst *nicht* bauen (erste Runde)

Ziele, Gewohnheits-Ketten mit Belohnungen, Stimmungs-Tracking, Teilen/Mehrbenutzer, Mobil-App, KI-Vorschläge. Jedes davon verwässert die These. Erst wenn „Heute" und „Rückblick" täglich benutzt werden.

---

## Screens für die erste Design-Runde

Vier Screens plus Launcher. Bewusst knapp gehalten — vier gut durchgestaltete Screens bringen dir mehr als sieben halbe.

1. `index.html` — Übersicht/Launcher, verlinkt alle Screens
2. `heute.html` — **der Anker**, inkl. geöffneter Capture-Palette als Zustand
3. `projekt-detail.html` — Projekt mit Markdown-Notiz, Aufgaben, Zeitverlauf
4. `rueckblick.html` — Wochenauswertung, der Payoff-Screen
5. `habits.html` — Heatmap und Historie

`notizen.html` und `kalender.html` folgen in Runde zwei, sobald das Muster steht.

---

## Visuelle Richtung

**Haltung:** Werkzeug, kein Dashboard. Nah am Terminal, aber mit der Ruhe eines gut gesetzten Buchs. Zurückhaltend genug, dass man es jeden Tag ansieht.

- **Fläche:** Dunkel, wie gehabt. Fast-Schwarz als Seite, eine Stufe heller für erhöhte Flächen, Haarlinien statt Schatten zur Trennung. Keine Verläufe.
- **Akzent:** Genau einer — das Teal/Grün, das in deiner Rail schon steckt. Maximal zweimal pro Screen: aktiver Nav-Zustand + die eine primäre Aktion. Zustände (erledigt, überfällig, verpasst) laufen über Helligkeit und Deckkraft, nicht über zusätzliche Farben.
- **Typografie, drei Rollen:**
  - *Serif* für Screen-Titel und die großen Zahlen im Rückblick — gibt dem lifelog Tagebuch-Charakter und trennt es von jedem dunklen Terminal-Klon.
  - *Sans* für Aufgabentexte, Notizen, Fließtext.
  - *Mono* für Datum, Dauer, Tags, Streaks — genau wie jetzt, aber **leiser**: kleiner, gedämpft, rechtsbündig. Metadaten stützen, sie konkurrieren nicht.
- **Dichte:** Zeilen, keine Karten. Der Karten-Rahmen aus dem Screenshot verschwindet meist — Abschnittsüberschrift plus Haarlinie reicht und spart die halbe visuelle Last.
- **Zustände:** Hover hebt den Zeilenhintergrund leicht an (nie den Text abdunkeln). Sichtbarer Fokus-Ring auf allem, was per Tastatur erreichbar ist. Erledigte Aufgaben werden gedämpft, nicht durchgestrichen und weggeworfen.
- **Ein Kunstgriff, nicht drei:** die Tagesachse auf „Heute" — eine durchgehende vertikale Linie mit „jetzt"-Markierung, die Termine, Aufgaben und Habit-Ticks auf derselben Zeitachse zusammenführt. Das ist das Bild, an das man sich erinnert.

---

## Interaktionsregeln

- **Cmd/Ctrl+K** öffnet Capture von überall. Ein Feld, dieselbe Syntax wie jetzt.
- **Syntax:** `#tag` · `@datum` · `15m`/`2h` Dauer · `+projekt` · `!` Priorität. Live-Vorschau parst und zeigt farbig, was erkannt wurde — das macht die Syntax lernbar, ohne Hilfetext.
- Eine Zeile ist überall klickbar (öffnet Detail), die Checkbox schaltet ohne Navigation um.
- **Empty States tragen Inhalt:** „Nichts für heute" zeigt die Capture-Syntax als Beispiel, nicht ein graues Icon.
- Tastatur zuerst: `j`/`k` durch Zeilen, `x` abhaken, `e` bearbeiten.

---

## Offene Fragen (bitte hier direkt beantworten)

- [x] **Zeit:** Sind die `2h` eine *Schätzung* oder *gemessene* Zeit? → Bei gemessen braucht „Heute" einen Timer-Zustand und Projekte einen Zeitverlauf. (Geschaetzte Zeit weil mein Ziel ist es, dass man so eien ansicht hat mit der man die Aufgaben fuer heute nimmt und mann dann die aneinander reiht und amn sieht wann man fertig waere und wie lange es wirklich dauert)
- [x] **Kalender:** Nur ICS-Import (lesend) oder zwei Wege (Google/CalDAV, schreibend)? → Entscheidet, ob der Kalender eine Anzeige oder eine bearbeitbare Oberfläche ist.(Schon mit Kalender einbinfung kommt aber spaeter)
- [x] **Habits:** Binär (erledigt / nicht erledigt) oder mit Menge (z. B. „30 min gelesen")? → Ändert Habit-Streifen und Rückblick-Auswertung.(Beides sollte moeglich sein)
- [x] **Betrieb:** Nur du, lokal? → Falls ja, entfallen Konto-, Sync- und Freigabe-Elemente komplett.(Stand jetzt nur ich. Villeicht publise ich es so als self deployed ding aber stand jetzt nur fuer mich selbst)
- [x] **Mobil:** Nur Desktop oder auch unterwegs erfassen? → Bei mobil braucht Capture einen eigenen kompakten Zustand.(Eigentlich reicht nur Laptop weil ich bin eh immer dahinter )

*Meine Annahme, falls du nichts änderst: gemessene Zeit, ICS lesend, binäre Habits, lokal und einbenutzerig, Desktop-first.*

---

## Nächster Schritt

Lies den Plan durch und ändere direkt in dieser Datei, was nicht passt — besonders die drei BLUF-Entscheidungen und die Verteilung „Heute vs. ausgelagert". Danach wechsle in den Design-Modus und sag mir, dass ich aus `lifelog-design-plan.md` bauen soll. Ich fange dann mit `heute.html` an, weil dieser Screen die These trägt; die übrigen leiten sich daraus ab.
