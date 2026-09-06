# htmx-Port — Fragmentgrenzen und Endpunkte

Der Prototyp ist bewusst so gebaut, dass der Port mechanisch ist: **jede Region mit
einer `id` ist ein Fragment.** Das Markup innerhalb dieser `id` ist exakt das, was dein
Server zurückgeben muss — nicht mehr, nicht weniger.

## Reihenfolge beim Port

1. `assets/lifelog.css` **unverändert übernehmen.** Keine Klasse ändert sich beim Port.
2. Jede `.html` in ein Server-Template überführen. Die Rail und die Capture-Leiste sind
   Layout-Bestandteile, keine Fragmente — die kommen ins Basis-Template.
3. Die `hx-*`-Attribute aus den HTML-Kommentaren an die genannten Elemente schreiben.
4. `assets/lifelog.js` **löschen.** Jede Funktion darin hat unten einen Endpunkt-Ersatz.

## Fragmente und Endpunkte

| Fragment (`id`) | Datei | Endpunkt | Trigger | Swap |
|---|---|---|---|---|
| `screen-head` | heute.html | `GET /heute/head?d=…` | Tageswechsel | `outerHTML` |
| `jetzt` | heute.html | `GET /heute/jetzt` | `every 60s`, nach Capture | `outerHTML` |
| `habit-strip` | heute.html | `GET /habits/heute` | nach Tick | `outerHTML` |
| `tagesachse` | heute.html | `GET /heute/achse` | `every 60s` | `outerHTML` |
| `capture-preview` | alle | `POST /capture/parse` | `keyup changed delay:120ms` | `innerHTML` |
| `wochen-zahlen` | rueckblick.html | `GET /rueckblick/{kw}/zahlen` | `load`, Wochenwechsel | `outerHTML` |
| `wochen-bilanz` | rueckblick.html | `GET /rueckblick/{kw}/bilanz` | Wochenwechsel | `outerHTML` |
| `wochen-balken` | rueckblick.html | `GET /rueckblick/{kw}/balken` | Wochenwechsel | `outerHTML` |
| `woche` | habits.html | `GET /habits/woche?kw=2026-W36` | nach Tick, Wochenwechsel | `outerHTML` |
| `habit-historie` | habits.html | `GET /habits/historie?weeks=…` | Zeitraum-Umschaltung | `outerHTML` |
| `projekt-panel` | projekt-detail.html | `GET /projekte/{slug}/{tab}` | Tab-Klick, `hx-push-url` | `innerHTML` |
| `projektliste` | projekte.html | `GET /projekte` | `load` | `outerHTML` |
| `notizen` | notizen.html | `GET /notizen?src=…&q=…` | Quelle, Suche `delay:200ms` | `outerHTML` |
| `kalender` | kalender.html | `GET /kalender?kw=2026-W36` | Wochenwechsel, `every 60s` | `outerHTML` |
| `git-jahr` | git.html | `GET /git?jahr=2026` | Jahreswechsel, `hx-push-url` | `outerHTML` |
| `projekt-fakten` | projekt-detail.html | `GET /projekte/{slug}/fakten` | `load`, `taskChanged from:body` | `outerHTML` |

## Aktionen

| Element | Endpunkt | Ziel |
|---|---|---|
| `.check` (Aufgabe abhaken) | `POST /tasks/{id}/toggle` | `hx-target="closest .task"` · `outerHTML` |
| `.habit` (Habit ticken) | `POST /habits/{slug}/tick?date=…` | `hx-target="#habit-strip"` · `outerHTML` |
| `.day` (Tag im Wochenraster) | `POST /habits/{slug}/tick?date=…` | `hx-target="#woche"` · `outerHTML` |
| `#habit-form` (Dialog speichern) | `POST /habits/{slug}` bzw. `POST /habits` | `hx-target="#woche"` · `outerHTML` |
| Archivieren | `POST /habits/{slug}/archive` | `hx-target="body"` · `hx-push-url="false"` |
| `.tree-item` (Finder) | `GET /notizen?src=…` bzw. `/projekte/{slug}/datei/{pfad}` | `hx-target=".finder-main"` · `innerHTML` |
| `.kb-card` (Board) | `POST /tasks/{id}/status` | `hx-target="#kanban"` · `outerHTML` |
| `.kb-add` | `POST /tasks?status=…&projekt=…` | `hx-target="closest .kb-list"` · `beforeend` |
| `[data-theme-toggle]` | rein clientseitig | `localStorage`, kein Request |
| `form.capture` (Anlegen) | `POST /capture` | `hx-target="#jetzt"` · `outerHTML` |
| `#daylog` (Tageslog) | `POST /log/{datum}` | `hx-trigger="input changed delay:800ms"` · `hx-swap="none"` |
| `#wochennotiz` | `POST /rueckblick/{kw}/notiz` | `hx-trigger="input changed delay:800ms"` · `hx-swap="none"` |
| `.stepper` (Tag / Woche) | `GET /heute?d=…` bzw. `/rueckblick/{kw}` | `hx-target="body"` · `hx-push-url="true"` |

## Der Git-Screen zieht aus der GitHub-API

Der Prototyp hält beide Jahre im Markup und blendet um; der Server liefert eins.
Vier Werte je Jahr genügen für den ganzen Aufmacher — `total`, `activeDays`,
`bestDay` und `longestStreak`. Der Rest ist Rechnung:

```
Ø je aktivem Tag      = total / activeDays
Anteil bester Tag     = bestValue / total
Vergleich zum Vorjahr = total / totalVorjahr
```

Die Tagesmatrix kommt aus `GET /users/{login}/events` bzw. der GraphQL-Query
`contributionsCollection.contributionCalendar` — **im Prototyp ist sie aus den
echten Jahreswerten rekonstruiert**: Summe, aktive Tage, bester Tag und längste
Serie stimmen exakt, die Verteilung dazwischen ist gesetzt. Repo- und
PR-Listen sind echte Daten aus den Screenshots.

Ein ruhiges Jahr braucht denselben Aufbau, nur andere Sätze. 2025 zeigt das:
statt einer fast leeren Heatmap trägt der Monatsverlauf die Aussage
(„62 % in November und Dezember"). Nie „nur 82 Beiträge" schreiben.

## Das Habit-Modell gehört in den Server

Ein Habit hat einen **Rhythmus**, und der entscheidet über alles Weitere:

```
rhythm = daily                 -> jeder Tag ist geplant
       | fixed [0..6]          -> nur diese Wochentage sind geplant
       | flex  n               -> kein Tag ist geplant, n pro Woche ist das Ziel
```

Daraus folgen drei Zustände je Tag, und nur der Server kann sie korrekt bilden:

| Zustand | Bedingung |
|---|---|
| `done` | Eintrag vorhanden |
| `open` | geplant, Tag ist heute oder liegt in der Zukunft |
| `missed` | geplant, Tag liegt in der Vergangenheit, kein Eintrag |
| `off` | nicht geplant — **zählt weder als Lücke noch gegen die Serie** |

Wichtig für `GET /habits/woche` und die Historie: `off` ist kein Fehlschlag.
Quote rechnet `erledigt / geplant`, bei `flex` gegen `n × Wochen`. Die Serie
überspringt ungeplante Tage, statt an ihnen zu brechen — sonst reißt jede
Gym-Routine am Wochenende ab.

Ein Klick auf einen `off`-Tag trägt trotzdem ein (unplanmäßige Einheit). Der
Server antwortet mit dem neuen Rasterzustand; die Kachel entscheidet nicht selbst.

## Zwei Dinge, die du im Server lösen musst

**Mehrere Fragmente auf eine Aktion.** Ein abgehakter Task ändert die Zeile *und* den
Zähler `data-count-open` *und* die Tagesachse. Zwei saubere Wege:

- `hx-swap-oob="true"` auf den Nebenfragmenten in derselben Antwort, oder
- den Response-Header `HX-Trigger: taskChanged` setzen und die betroffenen Fragmente per
  `hx-trigger="taskChanged from:body"` selbst nachladen.

Der zweite Weg ist sauberer, wenn mehr als zwei Regionen betroffen sind.

**Der Capture-Parser gehört auf den Server.** Die Regeln in `lifelog.js` sind nur eine
Vorschau der Grammatik:

```
#tag        ein oder mehrere
+projekt    höchstens einer
@datum      heute | morgen | übermorgen | TT.MM.JJJJ
15m / 2h    Dauer
!           Priorität
```

Alles, was nach dem Entfernen dieser Token übrig bleibt, ist der Titel. Ein Eintrag ohne
Dauer und ohne Datum wird als Notiz angelegt, sonst als Aufgabe. Parse serverseitig und
gib die Chips als HTML zurück — dann bleibt die Grammatik an einer Stelle.

## Fortschrittsanzeige

`hx-indicator` ist im Prototyp noch nicht gestaltet. Empfehlung: keine Spinner, sondern
`opacity: 0.55` auf dem tauschenden Fragment über die vorhandene `.htmx-request`-Klasse.
Die Screens sind ruhig genug, dass ein Spinner als Störung auffällt.
