---
# bt-362n
title: E10 — Tag-Management-Page (zentrale Tag-Definition)
status: todo
type: epic
priority: high
tags:
    - to-review
created_at: 2026-07-16T15:39:48Z
updated_at: 2026-07-16T22:47:53Z
---

E10 — Tag-Management-Page (zentrale Tag-Definition). Liefert das Feature aus
`bt-6oyy`: eine eigene, im Command-Center erreichbare Page zur zentralen
Tag-Verwaltung (definieren/anlegen, umbenennen mit Propagation über alle
Beans, entfernen), plus Suggest-Mode-Integration in den bestehenden
Tag-Picker (`t`, `box_picker_tag.go`) — definierte Tags werden dort
priorisiert angeboten, freie Tags bleiben weiterhin erlaubt (kein strict
mode). Realisierungsplan: `docs/plans/tag-management/epic-E10-plan.md`
(volle Herleitung, Code-Sketches, Golden-/tmux-Vorgaben je Task). Dieses
Bean trägt den GETEILTEN Kontext (DRY) — Task-Bodies sind self-contained,
zitieren aber diesen Body als Design-Quelle statt ihn zu duplizieren.

## Auftragsquelle

`bt-6oyy` (Feature-bean, PO-Entscheide 2026-07-16): Kette startet direkt
nach E9 (`bt-tct9`, jetzt to-review), eigenes Epos. Fixe PO-Vorgaben:
Persistenz repo-lokal (exakter Dateiname/Format = Planner-Entscheid,
s. D01-D04), Tag-Picker läuft im Suggest-Mode (definierte Tags priorisiert,
freie Tags bleiben erlaubt, kein strict mode).

## Empirische Verifikation (vor Task-Schnitt, wie gefordert)

**beans-Scan-Verhalten** (Referenz-Klon `~/Obsidian/tools/lean-stack/beans-src`,
`pkg/beancore/core.go:128-179` `loadFromDisk`/`loadBean`, UND live gegen
dieses Repo getestet, Datei danach restlos entfernt):
`filepath.WalkDir` überspringt jede Nicht-`.md`-Datei im `.beans/`-Baum
ohne Fehler (`if !strings.HasSuffix(d.Name(), ".md") { return nil }`,
core.go:151-154) UND dot-präfigierte Unterordner komplett
(`filepath.SkipDir`, core.go:144-149). Live-Probe (dieses Repo,
0.4.2-Binary): eine `.beans/zz-scratch-probe.yml`-Datei UND eine
absichtlich unstrukturierte `.beans/zz-scratch-probe.md`-Datei (kein
Frontmatter) wurden beide klaglos toleriert — `beans list --json` (79
Beans, unverändert) und `beans check` (`All checks passed`) blieben grün
in beiden Fällen; Dateien danach gelöscht, `git status --porcelain` leer.
**Trotzdem (D01-Rationale unten): die Registry-Datei liegt NICHT in
`.beans/`** — nicht weil es nötig wäre, sondern weil das Repo-Root ohnehin
der sauberere, von der beans-Autorität komplett entkoppelte Ort ist
(spiegelt `.beans.yml` selbst, das auch im Root liegt, nicht in `.beans/`).

**Tag-Rename-Propagation (CLI-Verben, `beans update --help`, 0.4.2):**
`--tag`/`--remove-tag` sind beide `stringArray` (wiederholbar, frei
kombinierbar in EINEM Aufruf — das nutzt `data.Client.SetTags` bereits
heute für Einzel-Bean-Diffs). Es gibt **keinen** Bulk-/Rename-Verb auf
CLI-Ebene (kein `--rename-tag`, keine Mehr-Bean-Transaktion) — ein
Tag-Rename über N Beans ist zwangsläufig N unabhängige
`beans update <id> --if-match <etag> --tag <neu> --remove-tag <alt>
--json`-Aufrufe, je einer pro betroffenem Bean, ohne Cross-Bean-Atomarität
(D13 unten regelt die Fehlerbehandlung: continue-on-error, keine
Transaktion).

**Verwendungszähler-Quelle:** bereits vorhanden, keine neue Infrastruktur
nötig — `collectTagCounts` (`internal/tui/box_picker_tag.go:56-79`) zählt
`idx.ByID`-weit über alle Bean-Tags; `data.Index.WithTag(tag)`
(`internal/data/index.go:102-115`) liefert bereits die Bean-Liste zu einem
Tag (Basis für den Rename-Sweep, D13).

## Design-Entscheidungen (D01-D15, Planner, verbindlich)

**D01 — Persistenzort/-format.** Neue Datei `.beans-tags.yml` im
Repo-Root (Sibling zu `.beans.yml`, ermittelt über `client.RepoDir` —
NICHT über eine neue Discovery-Logik, dieselbe Root gilt bereits).
NICHT in `.beans/` (s. Verifikation oben — nicht nötig, aber sauberer:
komplett entkoppelt von beans' eigenem Scan/Autorität, kein Risiko bei
künftigen beans-Upgrades). Format: minimal, `tags: [<name>, <name>, …]`
— eine flache, alphabetisch sortiert gespeicherte Liste (deterministe
Diffs). KEIN Farb-/Beschreibungsfeld (YAGNI — kein PO-Wortlaut dafür,
Tag-Farbe ist laut design-spec.md §4/§8 ohnehin Hash-aus-Name-abgeleitet,
nicht gespeichert).

**D02 — Lade-Semantik.** Tolerant-Missing, spiegelt
`internal/config/settings.go`s `LoadSettings`-Vertrag: fehlende Datei →
leere Registry (kein Fehler), korrupte YAML → leere Registry (nie
crashen). Synchroner Read (`os.ReadFile`, kein `tea.Cmd`/keine
Subprocess-Latenz wie bei beans-CLI-Aufrufen — ein lokaler Datei-Read ist
schnell genug für einen direkten Aufruf im Update-Pfad, mirrort
`config.LoadSettings()`s eigene synchrone Aufruf-Konvention).

**D03 — Reload/Liveness.** Registry wird bei JEDEM Page-Open frisch von
Platte gelesen (`openTagManagementPage()`), mirrort `openLobby()`s „lädt
Settings neu, falls seit Start geändert"-Konvention. KEINE Erweiterung
des bestehenden fsnotify-Watchers (der beobachtet nur `.beans/`) auf diese
Datei — YAGNI, vermeidet eine zweite Watcher-/Reload-Race-Fläche. Ein
paralleler Agent, der die Datei mid-Session ändert, wird erst beim
nächsten Page-Open sichtbar (akzeptierter, dokumentierter Trade-off,
identisch zu Lobbys eigenem Verhalten für `~/.config/beans-tui/config.yaml`).

**D04 — Paket-Zuordnung.** Neue Datei `internal/data/tagdefs.go`
(Methoden auf `*data.Client`: `LoadTagDefs() ([]string, error)`,
`SaveTagDefs([]string) error` — Repo-scoped I/O, gleiche Package-Ebene wie
`tags.go`s `ValidTagName`), NICHT `internal/config` (das ist
User-global-Scope `~/.config/beans-tui/`, eine andere Autorität als
repo-lokale, teamfähige Daten). Reine Helfer `AddTagDefName`/
`RemoveTagDefName`/`RenameTagDefName([]string, ...) []string` (klein,
pure, TDD-freundlich) daneben.

**D05 — Einstiegspunkt.** Neuer `viewTagManagement` (viewID, Sibling zu
Backlog/Lobby). Erreichbar AUSSCHLIESSLICH über das Command-Center („go to
tags", mirrort den bestehenden „go to settings"-Präzedenzfall — auch das
Settings-Form hat keine eigene Taste). Kein neues globales Binding
verbraucht (Tastenraum ist knapp, kein PO-Wortlaut verlangt eine
dedizierte Taste).

**D06 — Dispatch-/Capture-Modell.** `viewTagManagement` ist ein
FULL-CAPTURE-Zustand, geprüft an DERSELBEN `handleKey`-Stelle wie
`viewLobby` (früh, VOR den ctrl+k/?/p-Bare-Match-Checks) — NICHT über die
generische `keyNodeAction`/Tree/Backlog-Kette geroutet. Begründung
(verifiziert, `update.go:1019-1029`): `focusedBean()`s `default`-Zweig
fällt auf den TREE-Cursor zurück, wenn `m.view` keinen eigenen Case hat —
liefe die Page NICHT full-capture, würden globale Node-Action-Tasten
(s/t/a/r/c/d/e) still gegen ein STALE, unzusammenhängendes Bean feuern,
während der PO die Tag-Registry ansieht (exakt die Fehlerklasse, die
LESSONS-LEARNED bereits zweimal dokumentiert — „Exit-Pfad-Inventur",
„Lobby-Exit im Hauptfall unerreichbar"). Mirrort Lobbys eigenen
Präzedenzfall 1:1 (full-capture, `?`/`ctrl+k`/`p` erreichen sie nicht,
`esc` ist der Rückweg).

**D07 — Chrome.** Reuse `Chrome()`/`breadcrumb()`/`footer()`/
`renderPane()` (Backlog-Stil Einzel-Listen-Pane, NICHT Lobbys
Bespoke-Banner-Zentrierung — strukturell ist die Tag-Page eine Liste,
näher an Backlog als an der Lobby-Startseite). `GlobalHint` wird LEER
übergeben (nicht `renderBindings(globalBindings())`) — keine der 4
globalen Tasten funktioniert während die Page Capture hält, sie
anzuzeigen wäre ein irreführendes, nicht-funktionales Versprechen (exakt
die Bug-Klasse, die D06/PF-11 bereits einmal für stale Footer-Hints
gefixt hat). Die eigene Footer-Zone-3 trägt jede wirklich funktionierende
Taste.

**D08 — Layout.** EINZEL-Pane-Liste (kein Master-Detail-Split mit
„Beans zu diesem Tag"-Drilldown) — PO-Wortlaut verlangte „Liste
definierter Tags + Verwendungszähler", kein Drilldown. Dokumentierter,
günstig nachrüstbarer YAGNI-Cut (`idx.WithTag(tag)` existiert bereits,
Fast-Follow trivial). `enter` auf einer Zeile ist ein dokumentierter
handled No-Op, reserviert für genau dieses Fast-Follow (mirrort Backlogs
eigenes enter-No-Op für eine flache Liste).

**D09 — Zeilen-Menge & Sortierung.** Die Liste zeigt die UNION aus (a)
jedem definierten Tag (Registry, auch mit Count 0) und (b) jedem Tag, der
aktuell auf irgendeinem Bean sitzt, aber NICHT in der Registry ist
(„frei"/undefiniert) — volle Sichtbarkeit, PO kann einen entdeckten
freien Tag direkt in die Registry heben. Zwei Gruppen: Definiert (alpha
sortiert — die Registry liest sich als vorhersagbarer Index) zuerst, dann
Undefiniert-in-Verwendung (Count absteigend — die meistgenutzten
unregistrierten Tags, die besten Beförderungs-Kandidaten, zuerst). Jede
Zeile trägt ein `defined bool`.

**D10 — Suggest-Mode im Tag-Picker.** `collectTagCounts`
(`box_picker_tag.go`) bekommt einen `defined map[string]bool`-Input,
Sortierung wächst um „defined" als NEUEN PRIMÄREN Schlüssel (definierte
Tags zuerst), der bestehende Count-absteigend/alpha-Tie-Break bleibt
SEKUNDÄR innerhalb jeder Gruppe. Registry wird frisch bei
`openTagPicker()` geladen (mirrort D03). Visuelle Unterscheidung: eine
IMMER reservierte Marker-Spalte vor jeder Zeile (PF-12-Konvention —
„immer reservierter Gutter, kein bedingtes Präfix, das das Layout
verschiebt", design-spec.md §15 PF-12) — definierte Tags ein gefülltes
Glyph, freie Tags ein gleich breiter Leerraum, Zeilen bleiben zwischen
den Gruppen immer bündig.

**D11 — Create-Semantik.** Gleiche Namens-Grammatik wie freie Tags heute
(`data.ValidTagName`, wörtlich wiederverwendet) + Dedupe gegen bestehende
Registry-Einträge (Inline-Fehler statt Duplikat, mirrort
`box_picker_tag.go`s eigene New-Tag-Input-Fehler-UX). Eine neue Definition
BERÜHRT KEIN Bean (kein Bean bekommt den Tag automatisch) — sie wird ab
sofort im Picker sichtbar/priorisiert. Mirrort design-spec.md §4 „Tags
entstehen implizit" — Definieren ist ein reiner Registry-Akt, entkoppelt
von Bean-Mutation.

**D12 — Delete-Semantik.** REGISTRY-ONLY-Entfernung — Beans, die den Tag
aktuell tragen, BEHALTEN ihn (er wird wieder ein „freier"/undefinierter
Tag, weiterhin voll funktional, Suggest-Mode erlaubt ihn weiterhin, nur
nicht mehr priorisiert). Begründung: der Auftragstext hängt die
„berührt jedes Bean"-Warnung EXPLIZIT nur ans Rename („rename müsste alle
Beans... umschreiben"), nie ans Entfernen — im Umkehrschluss ist Delete
der günstige, lokale, sichere Default; ein destruktiver
„Delete-und-überall-strippen"-Modus ist ein anderes, schwereres Feature,
das niemand verlangt hat (YAGNI, kein Bean angelegt, Fast-Follow nur auf
expliziten PO-Wunsch, s. Q01). Der Confirm-Dialog zeigt trotzdem den
LIVE-Verwendungszähler VOR dem Löschen (mirrort `box_confirm_delete.go`s
eigene Kinder-/Links-Zähler-Warnung), damit der PO informiert entscheidet
— auch wenn die Operation selbst nicht destruktiv ist.

**D13 — Rename-Semantik & Propagation.** Registry-Rename wird ZUERST
angewandt (reiner lokaler Datei-Op, kann praktisch nicht fehlschlagen) —
UNABHÄNGIG vom Ausgang des Bean-Sweeps. Danach läuft ein Best-Effort,
CONTINUE-ON-ERROR async Sweep (neuer `renameTagCmd`, EIN neuer
Message-Typ `tagRenameDoneMsg` — die zweite bewusste Ausnahme vom
geteilten `mutationDoneMsg`-Tail, mirrort `createDoneMsg`s eigenen
Präzedenzfall für einen reicheren Rückgabewert als nur EIN Error) über
`idx.WithTag(alt)`: EIN `data.Client.SetTags(id, add=[neu],
remove=[alt], etag)`-Aufruf PRO Bean (die BESTEHENDE SetTags-Methode,
KEINE neue Client-Methode für den Einzel-Bean-Call nötig — nutzt die
bereits etablierte Single-Etag-No-Cascade-Kombi-Diff-Konvention wieder).
Es gibt KEINE Cross-Bean-Transaktion in beans (verifiziert, `beans
update --help` kennt keinen Bulk-/Rename-Verb) — ein stales ETag auf Bean
K darf Bean K+1..N nicht abbrechen. Ergebnis (Anzahl umbenannt / Anzahl
fehlgeschlagen, erste Fehlermeldung) landet in EINEM Toast; `m.idx` lädt
danach neu (mirrort `applyMutationResult`s eigene „immer neu laden
danach"-Konvention), damit Picker/Tree den neuen Tag sofort überall
zeigen, wo er gelandet ist.

**D14 — Geteilter Freitext-Input-Submodus.** Create (T3) führt einen
page-lokalen Textinput-Submodus ein (`tagMgmtInputActive`/
`tagMgmtInputMode` „create"|„rename"/`tagMgmtInput`), mirrort
`box_picker_tag.go`s eigenen `tagInputActive`/`tagInput`-Präzedenzfall
(EIN dauerhaftes `textinput.Model`, reset+fokussiert bei Open, jede Taste
außer enter/esc gehört dem Input). Rename (T5) nutzt DENSELBEN Submodus
wieder (vorbefüllt mit dem alten Namen) statt einen zweiten, parallelen
Input-Mechanismus zu erfinden (mirrort E9s T5←T4-Präzedenzfall „ein
geteilter Helfer, zweiter Verbraucher erfindet ihn nicht neu").

**D15 — Delete-Confirm.** Ein page-lokales Bool+Ziel-Paar
(`tagMgmtDeleteConfirm`/`tagMgmtDeleteTarget`), KEIN neuer `overlayID`-Case
— mirrort `m.confirmQuit`s eigenen Präzedenzfall (types.go: „E2s eigene
Bools... bewusst NICHT ins overlayID-Enum zurückgeholt") für ein
einfaches, page-lokales Ja/Nein-Gate, das kein view-übergreifendes
Floating-Overlay ist.

## Offene Fragen (Q01-Q03, NICHT blockierend — Kette plant darum)

- **Q01:** Soll „Tag-Definition entfernen" optional auch einen
  destruktiven Modus anbieten, der den Tag zusätzlich von JEDEM Bean
  entfernt (GitHub-Label-Delete-Semantik), oder bleibt das immer eine
  manuelle/Rename-basierte Operation? D12 nimmt Registry-only (nicht
  destruktiv) als v1-Default an — echter Werturteil-Punkt über
  destruktives Default-Verhalten, den ich nicht sicher ableiten kann.
- **Q02:** Sollen definierte, aber aktuell unbenutzte Tags (Count 0)
  visuell/funktional als „Aufräum-Kandidat" markiert werden, oder ist das
  Rauschen? Nicht gebaut in v1 (D09 listet sie schlicht mit Count 0).
- **Q03:** Soll die BESTEHENDE „n"-Neuanlage im Tag-Picker selbst
  (`box_picker_tag.go`, B14/E8) künftig AUCH die Registry befüllen (ein
  Mental-Modell: „neuer Tag" = „definierter Tag"), statt wie heute nur
  den aktuellen Bean zu taggen? T6 lässt B14s bestehendes Verhalten
  UNVERÄNDERT (nur Sortierung/Anzeige wächst) — eine Verhaltensänderung an
  einem bereits abgenommenen, geschlossenen Feature (E8) wäre ein
  zusätzlicher Werturteil-Schritt außerhalb dieses Auftrags.

## Task-Übersicht

| Task | bean | Inhalt | blocked_by |
|---|---|---|---|
| T1 | siehe Plan-Dokument | Persistenzschicht `internal/data/tagdefs.go` | — |
| T2 | siehe Plan-Dokument | Page-Grundgerüst (viewTagManagement, read-only Liste) | T1 |
| T3 | siehe Plan-Dokument | Tag-Definition anlegen (Create + geteilter Input-Submodus) | T2 |
| T4 | siehe Plan-Dokument | Tag-Definition entfernen (Delete, registry-only) | T2 |
| T5 | siehe Plan-Dokument | Tag-Definition umbenennen + Propagation über alle Beans | T3 |
| T6 | siehe Plan-Dokument | Tag-Picker Suggest-Mode (collectTagCounts erweitert) | T1 |
| T7 | siehe Plan-Dokument | Abschluss (Voll-Validierung, design-spec.md, Epic to-review) | T2,T3,T4,T5,T6 |

Volle Herleitung, Akzeptanzkriterien, TDD-/Golden-/tmux-Vorgaben je Task:
`docs/plans/tag-management/epic-E10-plan.md`.

## Review-Merkposten aus T1 (2026-07-16, F04 — gilt für T3-T6)

SaveTagDefs/Add|Remove|RenameTagDefName validieren bewusst NICHT gegen ValidTagName (nur LoadTagDefs filtert defensiv) — Validierung gehört an die Eingabegrenze (D11, bean bt-49hh Notes). PFLICHT für jeden T3-T6-Reviewer: sicherstellen, dass jeder neue Aufrufer vor SaveTagDefs tatsächlich validiert.

## Q04 (2026-07-16, aus T5-Review F01 — PO-Frage, nicht blockierend)

Soll Rename eines definierten Tags AUF den Namen eines existierenden freien Tags erlaubt sein (= Merge: Registry-Eintrag umbenennen + Sweep vereinigt beide Tag-Populationen)? Aktuell per Dedupe abgelehnt (T3-Erbe, Dedupe gegen ALLE Zeilennamen). data.RenameTagDefName erlaubt es auf Datenebene bereits ('Rename onto an unregistered name is allowed'). Empfehlung Supervisor: v-nächste-Iteration, nicht in E10 — Merge hat eigene Confirm-UX-Fragen (Populations-Vereinigung irreversibel).


## US-Review 2026-07-16 (PO, Runde 7)

- US-16 (Command-Center 'go to tags'): accepted
- US-17 (Liste definiert+frei mit Marker/Count): accepted

## Nebenbefund (PO, Runde 7)

e/d wirkten auf Tag 'smoke' wie kaputte Keybinds -- Ursache: 'smoke' ist ein
FREIER Tag (Registry .beans-tags.yml existiert im Repo noch nicht, kein Tag
je definiert), e/d sind auf freien Zeilen laut D12/D13 bewusster No-Op. Echtes
Problem: Footer zeigt e/d unbedingt, ohne dass ersichtlich ist ob die
aktuelle Zeile definiert/frei ist -- kein Feedback bei No-Op-Tastendruck.
-> Bug bt-<folgt> angelegt.


## US-Review 2026-07-16 (PO, Runde 8)

- US-18 (n=neuer Tag, reiner Registry-Akt): accepted
- US-19 (d=Delete Registry-only, Live-Count-Confirm): accepted
- US-20 (e=Rename mit Bean-Propagation): accepted

## Nebenbefunde (PO, Runde 8)

- NB-5: 'n' auf einer bereits vorhandenen FREIEN Tag-Zeile registriert diesen
  Tag nicht -- PO erwartet: Cursor auf freier Zeile + n = genau dieser Tag
  wird automatisch registriert (adoptiert), statt nur ein Blank-Create-Input
  zu oeffnen. -> Feature bt-<folgt> angelegt.
- NB-6: e/d auf freier Zeile soll eine Notification zeigen, PO-Wortlaut:
  'unregistred tag - modification not possible' -- praezisiert bt-ct3k.


## US-Review 2026-07-16 (PO, Runde 9)

- US-21 (Tag-Picker Suggest-Mode, definiert zuerst): accepted

## US-Review Abschluss (2026-07-16)

21/21 User-Stories durchgesprochen: 20 accepted, 1 rejected (US-05, Feedback
-> bt-b0w0). 6 neue Follow-Up-beans aus Nebenbefunden: bt-9ipw (Feature,
Tag-Picker-Typeahead), bt-98cb (Bug, Accordion-Kollaps), bt-lg68 (Bug,
Datums-Dopplung Meta/History), bt-39cl (Bug, Tree-Children-Aufklapp-Fehler,
high), bt-ct3k (Bug, fehlendes Feedback e/d auf freier Zeile), bt-idm1
(Feature, n=Adopt auf freier Zeile).


## E11-Nacharbeit abgeschlossen (2026-07-17)

Alle E10-seitigen Items fertig, je mit APPROVED-Review: bt-ct3k (Toast bei e/d
auf freiem Tag, 2da5ef9) · bt-idm1 (n-Adopt direkt, b9ed10b + ValidTagName-
Fix-Runde d83c50a) · bt-9ipw (Tag-Picker Typeahead, df249d7 + Cursor-Test-
Fix-Runde 9a42af6). Nebenfund als eigenes bean: bt-l8e7 (Lobby i/k-Alias-Leak,
low). Voll-Suite auf integriertem main grün. PO-Abnahme steht aus (to-review).
