# lifelog — Roadmap

Persönliches Datenlager in Go. Eine Binary, SQLite, Collectors ziehen aus
externen Quellen in einen gemeinsamen Event-Store.

Themen, keine Schritte. Jeder Meilenstein sagt, was das Ding können soll.
Das Wie wird beim Bauen entschieden.

---

## Wozu das Ganze

**Wie verlief mein Tag, und was könnte ich besser machen?**

Rohe Zahlen beantworten das nicht. „12 Commits" ist weder gut noch schlecht,
solange kein Maßstab danebensteht. Jede Quelle weiß, *was* passiert ist.
Keine weiß, *was für ein Tag das war*. Apple Health kennt den Schlaf, aber
nicht, ob der Sonntag Urlaub war oder ein verkackter Arbeitstag.

**Projektgrenze** (2026-09-13):

> Lifelog erzeugt keine Daten, die eine Maschine hätte sammeln können.
> Es fragt nur nach dem, was kein Sensor misst.

Das schließt Aufgabenverwaltung aus und lässt zwei Dinge herein: **Kontext**
(was für ein Tag war das) und **Empfinden** (wie war er, wie erfolgreich).

**Plattformgrenze** (2026-09-13):

> Kein macOS-spezifischer Code außerhalb des jeweiligen Collector-Pakets.

M1–M4 laufen überall, wo Go läuft. Die Apple-Quellen sind Ergänzungen. Wer
sie nicht hat, bekommt eine Instanz ohne diese Collectors. Kein
Plugin-System, kein Interface auf Vorrat. Die Regel ist Disziplin.

**Ein Mensch pro Instanz.** Kein Login, keine Mandanten, keine geteilte
Datenbank.

---

## Was steht (September 2026)

Die GitHub-Strecke läuft end-to-end: eine Binary serviert Dashboard und
Git-Seite über htmx und serverseitig gerendertes SVG aus einem SQLite
Event-Store. Tages-Events und Jahres-Aggregate sind über `granularity`
unterscheidbar, Jahreswerte landen nicht mehr auf dem 1. Januar der
Tagesachse.

Dazu `/day`-Detailansicht, Backfill über vergangene Jahre, PR- und
Repo-Commit-Events. Todos wurden eingebaut, passten aber nicht zur
Grundvision des Projekts und sind wieder draußen (M2, 2026-09-14).

---

## M1 — Der Tag bekommt einen Maßstab

Ein Kontextwort, zwei Zahlen, optional eine Zeile Text.

Zwei Zahlen, weil sie Verschiedenes messen: **wie war der Tag** und **wie
erfolgreich war er**. Ein guter Tag kann unproduktiv sein. Beide werden
getippt und nie berechnet. Sie sind die Zielvariable, gegen die M7 alles
andere prüft.

Erfassung muss billiger sein als sie zu überspringen: ein Feld, ein Enter,
null Klicks. Keine Dropdowns, keine Pflichtfelder. Inline geparst, wie bei
den Todos — der Parser ist mit M2 gegangen und wird hier neu geschrieben.

Der Freitext ist **eine Zeile und bleibt es**. Nicht durchsuchbar, nicht
editierbar, keine Tags. Sonst ist es eine Notiz, und Notizen gehören nach
Obsidian.

**Ziele statt Note.** Selbstgesetzte Schwellen, als Häkchen angezeigt:
`Schlaf 6,2 h ✗ · Commits 9 ✓`. Nicht zu einem Prozentwert verrechnet, sonst
ist die Gewichtung durch die Hintertür wieder da. Vier Zahlen in der Config,
kein Editor.

**Der Anstoß gehört dazu.** Das Risiko ist nicht die Eingabedauer, sondern
das Vergessen. Ohne täglichen Anstoß bleibt die Tabelle leer und M3 und M7
sind wertlos. Welcher Weg, wird beim Bauen entschieden.

*Fertig, wenn:* ein Tag in unter zehn Sekunden beschrieben ist und abends
etwas daran erinnert.

---

## M2 — Todos raus ✓ (2026-09-14)

Aufgaben leben in Apple Reminders und Obsidian, und zwar besser.

Weg sind Store-Funktionen, Handler, Templates, Seite und Nav-Eintrag; die
Tabelle fällt über `00007_drop_todos.sql`. Migrationen 00003–00006 bleiben
stehen, goose-Historie wurde nicht umgeschrieben. Das Down von 00007 ist
bewusst ein `RAISE(FAIL)`: zurück geht es nicht.

`dashboard.Day(events, day)` trägt keine Todos mehr in der Signatur. Der
Tages-View wird um M1 herum neu gebaut.

Mit weg ist der Inline-Parser (`15m #tag @datum`). M1 baut ihn gegen seine
eigene Eingabe neu; die alte Fassung steht in der Historie.

---

## M3 — Wochenrückblick

Sieben Tage nebeneinander: Kontext, die beiden Bewertungen, Sensorwerte.
Rohe Zeilen, kein Modell, keine Aussage.

Muss **ab Tag 7** etwas zeigen. Eine ehrliche Auswertung braucht acht
Wochen, und an dieser Durststrecke ist das Projekt schon einmal erlahmt. Das
Muster erkennt man selbst, bevor der Rechner es sagen darf.

Dazu das Dashboard, das heute dieselbe Jahres-Heatmap wie `/git` rendert:
Jahresgitter raus, rollendes Fenster (heute / 7 / 30 Tage) rein.

---

## M4 — Ausgabe: Markdown, JSON, CSV

Ein Query, drei Formatter. `encoding/json` und `encoding/csv` sind stdlib.

**Markdown → Obsidian**, nie umgekehrt. Die Tagesnotiz bekommt die Messwerte
vorgefüllt. Lifelog liefert die Zahlen, Obsidian die Worte.

**CSV und JSON** für Tabellenkalkulation oder eine KI. CSV zuerst: weniger
Tokens, und ein Tabellenformat lädt zum Vergleichen ein.

Was hinausgeht, ist ein Lebensprofil. Der Default exportiert deshalb keinen
Freitext und keine Aufgabentitel, nur Kategorien und Zahlen. Alles andere
braucht ein Flag.

---

## M5 — Erinnerungen als Quelle

Lifelog liest, was abgehakt wurde: Text, Zeitpunkt, Liste.

Kein Widerspruch zu M2. Eine erledigte Erinnerung ist ein Messwert wie ein
Commit, sie ist bereits passiert. Lifelog legt nichts an.

Damit kommt die halbe Antwort auf M1 von der Maschine: fünf erledigte Dinge
aus der Liste *Arbeit*, also war es ein Arbeitstag. Das Wort muss niemand
tippen.

**Listen, nicht Tags.** Listen gibt es seit der ersten Version, jeder
Zugriffsweg gibt sie zuverlässig heraus. Tags sind neu und fehlen älteren
Schnittstellen.

Zugriff über macOS-Bordmittel (Kurzbefehl oder AppleScript), Ausgabe wird
gelesen. Kein Swift-Build, keine neue Abhängigkeit. Langsam, läuft aber im
Hintergrund und nie, während jemand auf eine Seite wartet. Falls es zu
langsam sein sollte, besteht die Möglichkeit, ein kleines
EventKit-Hilfsprogramm zu schreiben.

**Vorgezogen** (2026-09-13), weil die Listen bereits geführt werden. Der
Preis: vor M1 ist das eine dritte Spalte Zahlen ohne Maßstab. M1 kommt
direkt danach.

---

## M6 — Apple Health

Schlaf und Schritte. Erste Quelle, die nicht GitHub ist.

Der Export ist ein mehrere hundert MB großes XML und muss streamend gelesen
werden. Eine Nacht ist genau ein Event, datiert auf den Aufwachtag.

Vor dem Bauen zu klären: jede bisherige Quelle legt Rohdaten in
`raw_payloads` ab, und `InsertEvents` löscht per `raw_id` neu ein. Das ist
die einzige Idempotenz-Klammer. Ein 500-MB-XML passt dort nicht hinein, also
braucht Health einen eigenen Weg und eine eigene Antwort darauf, was beim
zweiten Import mit den Events des ersten passiert.

---

## M7 — Auswertung

Erst wenn genug Tage da sind. Gleiches mit Gleichem vergleichen:

> *Arbeitstage, die du gut fandest: Ø 7,4 h Schlaf, 9 Commits, 1 Training.
> Die schlechten: Ø 5,9 h, 3 Commits, 0 Training.*

**Hier entsteht der Maßstab, und nur hier.** Nicht gewichtet, sondern aus
deinen eigenen Bewertungen gelesen. Das ist die einzige Zahl im System, die
dir etwas sagen kann, das du nicht schon wusstest.

Die Schwelle steht vorher fest: unter 40 bewerteten Tagen oder weniger als
10 pro Gruppe wird keine Gegenüberstellung gezeigt, sondern gezählt, wie
viele fehlen. Und selbst dann ist das Ergebnis ein Hinweis, keine Messung.
Die Zielvariable ist selbst vergeben, die Stichprobe klein.

*Fertig ist es, wenn ein Satz herauskommt, nach dem man etwas anders macht.*

---

## M8 — Collector-Abstraktion

Erst wenn neben GitHub zwei weitere Quellen stehen. Drei Implementierungen
zeigen, was sie teilen. Eine zeigt nichts, und geraten wird hier nicht.

Sie sind bewusst verschieden: GitHub ist eine HTTP-API mit Token,
Erinnerungen ein lokaler Prozessaufruf, Health ein Datei-Import. Was diese
drei teilen, teilen alle.

Dazu gehört `cmd/lifelog/sync.go`: drei GitHub-Fetches hintereinander,
Normalisierung per `strings.HasPrefix` auf der `ExternalID`. Bei drei Quellen
wird das eine Kaskade in `main`.

Geplanter Sync statt manueller Flags, und eine fehlschlagende Quelle reißt
die anderen nicht mit. Kleine Version zuerst: `time.Ticker` alle sechs
Stunden, Fehler in `sync_state` und ins Log. Kein Backoff, keine Job-Queue,
solange nichts davon gemessen fehlt.

---

## Out of Scope

Alles hier gestrichen 2026-09-13.

- **Aufgabenverwaltung.** Lifelog legt nichts an und hakt nichts ab. Dass M5
  erledigte Erinnerungen *liest*, ist kein Widerspruch.
- **Berechnete Tagesnote.** Kein Score aus Schlaf, Commits und Aufgaben,
  weder fest verdrahtet noch konfigurierbar. Die Gewichte wären geraten, und
  man optimiert dann auf sie: 12 Commits bei 4 h Schlaf bekämen eine gute
  Note. Selbstgesetzte Gewichte geben nur das eigene Bauchgefühl als Messwert
  zurück. Stattdessen Schwellen in M1 und der gelesene Maßstab in M7.
- **Gedanken-Inbox, Notizen.** Gehört nach Obsidian und ist dort schneller
  erreicht. Lifelog müsste Editieren, Ordnen, Tags und Suche nachbauen und
  wäre ein schlechteres Obsidian. Die eine Zeile Freitext in M1 ist die
  Grenze, nicht der Anfang.
- **Obsidian-Collector, Backlinks.** Dateien scannen und Frontmatter parsen
  ist die Komplexität, die vermieden werden sollte.
- **Kalender.** Sollte das Kontextsignal liefern. Das tun ab M5 die
  Erinnerungslisten, und zwar aus dem, was passiert ist, statt aus dem, was
  geplant war.
- **Mehrbenutzer, Login, Cloud.** „Empfinden 7" ist zwischen zwei Menschen
  bedeutungslos. Kein Netzwerkeffekt, nur Angriffsfläche auf einem
  Lebensprofil.
- **Strava.** Ausdauertraining, das gerade nicht stattfindet. Eine Quelle,
  die nicht gefüttert wird, liefert keine Daten.
- **openGym.** Braucht VPS, Deployment, Tailscale, Passkeys und ein junges
  Ein-Mann-Projekt, das gepinnt werden muss. Hevy Lifetime ist billiger.

## Später, nicht gestrichen

**Hevy** — Krafttraining: Sätze, Gewicht, Wiederholungen. REST-API mit
API-Key, kein OAuth. Vor dem Kauf prüfen, ob der Lifetime-Plan den
API-Zugang einschließt und der Workout-Endpunkt Sätze einzeln liefert statt
nur Zusammenfassungen.

**Hardcover** — Gelesenes. GraphQL mit Token, dasselbe Muster wie GitHub.
Ehrlich eingeordnet: beantwortet „wie verlief mein Tag" kaum besser als der
Rest. Zuletzt.

Beide nach M8, damit sie die Abstraktion benutzen statt sie zu erfinden.

---

## Mögliche Probleme

**Tagesgrenze.** Ein Tag läuft 04:00–03:59 Ortszeit. Ohne diese Regel zählt
ein Commit um 2 Uhr zum nächsten Tag, obwohl er zum Abend davor gehört. Dann
vergleicht M7 Tage, die es nie gab. Gilt genauso für Schlaf (Aufwachtag) und
Erinnerungen.

**Zeitzonen.** Zeitpunkte in UTC, `local_date` immer explizit daneben. Sonst
verschieben Reisen und Sommerzeit rückwirkend, welcher Tag welche Daten hat.

**Duplikate beim Re-Sync.** `UNIQUE (source, external_id)` plus Upsert. Ohne
das wächst jede Quelle bei jedem Sync.

**Rate Limits.** Alle sechs Stunden syncen, nicht alle fünf Minuten. Ein
gesperrter Token kostet mehr, als aktuelle Zahlen wert sind.

## Bekannte Mängel

- Die Git-Heatmap rendert bis zum 31. Dezember. Zukünftige Tage sehen aus
  wie Tage ohne Aktivität, ~40 % der Fläche bleibt leer.
- PR-Statuspunkte sind alle grau. Der `state` wird geholt und dekodiert,
  landet aber nie im `meta`. Re-Normalisieren aus den gecachten
  `raw_payloads` reicht.
- Der Design-Prototyp ist fertig, aber nicht in die Templates portiert.
- `internal/store/events.go` hat GitHub fest verdrahtet: `ListRepoCommits`,
  `ListPullRequests` und `EventYearRange` filtern auf `source = 'github'` und
  dekodieren GitHub-Meta im Store. Pro Quelle kämen zwei fast identische
  Funktionen dazu. Ausweg: generisches `ListEvents` plus Dekodieren im
  Handler. Fällig mit M8.
- `ListPullRequests` deklariert `meta` außerhalb der Zeilenschleife
  (`events.go:255`).
- Die Oberfläche ist auf Deutsch. Neue Templates deshalb auf Englisch, sonst
  wird die Umstellung mit jedem Template teurer.

---

## Layout

Ein Go-Modul, Monorepo. Go im Repo-Root, kein `backend/`.

```
cmd/lifelog/        Flags, Startup, HTTP-Server
internal/api/       Handler und Seiten
internal/store/     SQLite, Migrationen, Queries
internal/config/    Env und Flags
internal/github/    erste Quelle
internal/collector/ gemeinsame Typen
web/                Templates und statische Dateien
DesignPrototype/    Design-Vorlage (generiert)
```

Drin: `modernc.org/sqlite`, `pressly/goose`, `joho/godotenv`. Bewusst
draußen: HTTP-Framework, ORM, PostgreSQL, Redis, Message Queue,
Auth-Framework, npm.

## Sicherheit

Diese Datenbank ist ein präzises Profil eines Lebens.

- Bindung auf `127.0.0.1`.
- Fernzugriff nur über Tailscale oder WireGuard, keine Portweiterleitung.
- Secrets in `.env`, nie committet. Fremde API-Keys verschlüsselt ablegen.
- Datenbank nach `~/.local/share/lifelog/`, heute schützt sie nur
  `.gitignore`.
- Backup beim Start. Die `raw_payloads` sind reproduzierbar, die selbst
  erfassten Tage nicht.

---

Nach jedem Meilenstein committen. Der Ertrag kommt erst in Jahren,
sichtbarer Fortschritt hält das Projekt am Leben.
