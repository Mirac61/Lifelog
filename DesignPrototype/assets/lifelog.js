/* ─────────────────────────────────────────────────────────────────────
   lifelog — Prototyp-Interaktivitaet.

   NUR FUER DEN PROTOTYPEN. Beim htmx-Port ersetzt du jeden Block hier
   durch ein hx-post/hx-get auf deinen Server. Die Kommentare nennen den
   jeweils passenden Endpunkt. Danach kann diese Datei geloescht werden.
   ───────────────────────────────────────────────────────────────────── */
(function () {
  'use strict';

  /* ── Aufgabe abhaken ──────────────────────────────────────────────
     htmx: <button class="check" hx-post="/tasks/{id}/toggle"
                   hx-target="closest .task" hx-swap="outerHTML"> */
  document.addEventListener('click', function (e) {
    var check = e.target.closest('.check');
    if (!check) return;
    var task = check.closest('.task, .tl-item');
    if (!task) return;
    var done = task.getAttribute('data-done') === 'true';
    task.setAttribute('data-done', done ? 'false' : 'true');
    check.setAttribute('aria-checked', done ? 'false' : 'true');
    recountOpen();
  });

  /* ── Habit umschalten ─────────────────────────────────────────────
     htmx: <button class="habit" hx-post="/habits/{slug}/tick?date=2026-09-06"
                   hx-target="#habit-strip" hx-swap="outerHTML"> */
  document.addEventListener('click', function (e) {
    var habit = e.target.closest('.habit');
    if (!habit) return;
    var done = habit.getAttribute('data-done') === 'true';
    habit.setAttribute('data-done', done ? 'false' : 'true');
    habit.setAttribute('aria-pressed', done ? 'false' : 'true');
  });

  /* ── Offene Zeit im Pane-Kopf nachziehen ─────────────────────────
     Serverseitig kommt das mit dem Fragment mit — hier lokal gerechnet. */
  function recountOpen() {
    var out = document.querySelector('[data-count-open]');
    if (!out) return;
    var mins = 0, n = 0;
    document.querySelectorAll('.task[data-minutes]').forEach(function (t) {
      if (t.getAttribute('data-done') === 'true') return;
      mins += parseInt(t.getAttribute('data-minutes'), 10) || 0;
      n += 1;
    });
    var h = Math.floor(mins / 60), m = mins % 60;
    var time = h ? (m ? h + ' h ' + m + ' min' : h + ' h') : m + ' min';
    out.textContent = n === 0 ? 'alles erledigt' : n + ' offen · ' + time;
  }
  recountOpen();

  /* ── Capture: Live-Vorschau des geparsten Eintrags ────────────────
     htmx: <input hx-post="/capture/parse" hx-trigger="keyup changed delay:120ms"
                  hx-target="#capture-preview">
     und   <form hx-post="/capture" hx-target="#jetzt" hx-swap="outerHTML"> */
  var input = document.getElementById('capture-input');
  var preview = document.getElementById('capture-preview');

  var MONTHS = ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'];

  function parse(raw) {
    var rest = raw;
    var out = { tags: [], project: null, duration: null, date: null, prio: false };

    rest = rest.replace(/\+([\wäöüÄÖÜß-]+)/g, function (_, p) { out.project = p; return ''; });
    rest = rest.replace(/#([\wäöüÄÖÜß-]+)/g, function (_, t) { out.tags.push(t); return ''; });
    rest = rest.replace(/(\d+)\s?(min|m|h)\b/gi, function (_, n, u) {
      out.duration = u.toLowerCase() === 'h' ? n + ' h' : n + ' min';
      return '';
    });
    rest = rest.replace(/@(\S+)/g, function (_, d) { out.date = normalizeDate(d); return ''; });
    rest = rest.replace(/(^|\s)!(\s|$)/g, function () { out.prio = true; return ' '; });

    out.title = rest.replace(/\s+/g, ' ').trim();
    return out;
  }

  function normalizeDate(d) {
    var low = d.toLowerCase();
    if (low === 'heute') return 'So 6. Sep';
    if (low === 'morgen') return 'Mo 7. Sep';
    if (low === 'übermorgen' || low === 'uebermorgen') return 'Di 8. Sep';
    var m = low.match(/^(\d{1,2})\.(\d{1,2})\.?(\d{4})?$/);
    if (m) return parseInt(m[1], 10) + '. ' + (MONTHS[parseInt(m[2], 10) - 1] || m[2]);
    return d;
  }

  function chip(text, cls) {
    var s = document.createElement('span');
    s.className = 'chip' + (cls ? ' ' + cls : '');
    s.textContent = text;
    return s;
  }

  function render() {
    if (!preview) return;
    preview.textContent = '';
    var raw = input.value.trim();
    if (!raw) {
      preview.textContent = 'Erkannt wird: #tag · +projekt · @datum · 15m / 2h · ! für Priorität';
      return;
    }
    var p = parse(raw);
    preview.appendChild(chip(p.duration || p.date ? 'Aufgabe' : 'Notiz', 'chip-type'));
    if (p.title) preview.appendChild(chip(p.title.length > 42 ? p.title.slice(0, 42) + '…' : p.title));
    if (p.duration) preview.appendChild(chip(p.duration));
    if (p.date) preview.appendChild(chip(p.date));
    if (p.project) preview.appendChild(chip('+' + p.project));
    p.tags.forEach(function (t) { preview.appendChild(chip('#' + t)); });
    if (p.prio) preview.appendChild(chip('Priorität'));
  }

  if (input) {
    input.addEventListener('input', render);
    render();

    var form = input.closest('form');
    if (form) {
      form.addEventListener('submit', function (e) {
        e.preventDefault();          // htmx: hx-post="/capture" uebernimmt das
        if (!input.value.trim()) return;
        input.value = '';
        render();
        input.focus();
      });
    }
  }

  /* ── Cmd/Ctrl+K oeffnet Capture von ueberall ─────────────────────── */
  document.addEventListener('keydown', function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      if (input) { input.focus(); input.select(); }
      return;
    }
    if (e.key === 'Escape' && document.activeElement === input) { input.blur(); }
  });

  /* ── Tastatur: j / k durch Zeilen, x zum Abhaken ─────────────────── */
  var cursor = -1;
  document.addEventListener('keydown', function (e) {
    var tag = (document.activeElement && document.activeElement.tagName) || '';
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    var rows = Array.prototype.slice.call(document.querySelectorAll('.task'));
    if (!rows.length) return;

    if (e.key === 'j' || e.key === 'k') {
      e.preventDefault();
      cursor = e.key === 'j'
        ? Math.min(cursor + 1, rows.length - 1)
        : Math.max(cursor - 1, 0);
      var btn = rows[cursor].querySelector('.check');
      if (btn) btn.focus();
    } else if (e.key === 'x' && cursor >= 0) {
      e.preventDefault();
      var c = rows[cursor].querySelector('.check');
      if (c) c.click();
    }
  });

  /* ── Tabs (Projekt-Detail) ───────────────────────────────────────
     htmx: <button class="tab" hx-get="/projekte/lifelog/notiz"
                   hx-target="#projekt-panel" hx-swap="innerHTML"> */
  document.querySelectorAll('[data-tabs]').forEach(function (group) {
    group.addEventListener('click', function (e) {
      var tab = e.target.closest('.tab');
      if (!tab) return;
      group.querySelectorAll('.tab').forEach(function (t) {
        var on = t === tab;
        t.setAttribute('aria-selected', on ? 'true' : 'false');
        var panel = document.getElementById(t.getAttribute('aria-controls'));
        if (panel) panel.hidden = !on;
      });
    });
  });

  /* ── Wochenraster: Tag ein-/austragen ─────────────────────────────
     htmx: <button class="day" hx-post="/habits/{slug}/tick?date=YYYY-MM-DD"
                   hx-target="#woche" hx-swap="outerHTML">
     Der Server kennt den Plan und liefert den korrekten Folgezustand
     zurück — offen, ausgelassen oder nicht geplant. */
  function weekRow(cell) {
    var first = cell;
    while (first.previousElementSibling &&
           !first.previousElementSibling.classList.contains('week-name')) {
      first = first.previousElementSibling;
    }
    var days = [], cur = first;
    while (cur && !cur.classList.contains('week-tally')) {
      if (cur.classList.contains('day')) days.push(cur);
      cur = cur.nextElementSibling;
    }
    return { days: days, tally: cur };
  }

  function retally(cell) {
    var row = weekRow(cell);
    if (!row.tally) return;
    var done = row.days.filter(function (d) { return d.getAttribute('data-state') === 'done'; }).length;
    var plan = parseInt((row.tally.textContent.split('/')[1] || '0'), 10);
    row.tally.textContent = done + '/' + plan;
    row.tally.setAttribute('data-hit', done > plan ? 'over' : (done >= plan ? 'true' : 'false'));
  }

  document.addEventListener('click', function (e) {
    var day = e.target.closest('.day');
    if (!day) return;
    if (!day.hasAttribute('data-state0')) day.setAttribute('data-state0', day.getAttribute('data-state'));
    var back = day.getAttribute('data-state0');
    if (back === 'done') back = day.hasAttribute('data-today') ? 'open' : 'missed';
    day.setAttribute('data-state', day.getAttribute('data-state') === 'done' ? back : 'done');
    retally(day);
  });

  /* ── Dialog: Rhythmus, Ziel, Tageszeit ────────────────────────────
     Reines Markup. Beim Port wird das <form> zu
     hx-post="/habits/{slug}" hx-target="#woche" hx-swap="outerHTML". */
  var sheet = document.getElementById('habit-sheet');

  function setRhythm(kind) {
    if (!sheet) return;
    sheet.querySelectorAll('[data-rhythm] button').forEach(function (b) {
      b.setAttribute('aria-pressed', b.getAttribute('data-kind') === kind ? 'true' : 'false');
    });
    sheet.querySelectorAll('[data-when-kind]').forEach(function (f) {
      f.hidden = f.getAttribute('data-when-kind') !== kind;
    });
  }

  function openSheet(src) {
    if (!sheet) return;
    var isNew = !src;
    sheet.querySelector('#sheet-title').textContent = isNew ? 'Neuer Habit' : 'Habit bearbeiten';
    sheet.querySelector('#hb-name').value     = isNew ? '' : src.getAttribute('data-name');
    sheet.querySelector('#hb-goal').value     = isNew ? '' : src.getAttribute('data-goal');
    sheet.querySelector('#hb-per-week').value = isNew ? 3  : src.getAttribute('data-per-week');
    sheet.querySelector('#hb-when').value     = isNew ? 'egal' : src.getAttribute('data-when');
    sheet.querySelector('[data-od-id="btn-archivieren"]').hidden = isNew;

    var days = isNew ? [] : (src.getAttribute('data-days') || '').split(',').filter(Boolean);
    sheet.querySelectorAll('[data-dow-group] button').forEach(function (b) {
      b.setAttribute('aria-pressed', days.indexOf(b.getAttribute('data-dow')) > -1 ? 'true' : 'false');
    });
    setRhythm(isNew ? 'daily' : src.getAttribute('data-kind'));
    sheet.showModal();
    sheet.querySelector('#hb-name').focus();
  }

  document.addEventListener('click', function (e) {
    if (e.target.closest('[data-edit-habit]')) { openSheet(e.target.closest('[data-edit-habit]')); return; }
    if (e.target.closest('[data-new-habit]'))  { openSheet(null); return; }
    if (e.target.closest('[data-close-sheet]') && sheet) { sheet.close(); return; }

    var seg = e.target.closest('[data-rhythm] button');
    if (seg) { setRhythm(seg.getAttribute('data-kind')); return; }

    var dow = e.target.closest('[data-dow-group] button');
    if (dow) { dow.setAttribute('aria-pressed', dow.getAttribute('aria-pressed') === 'true' ? 'false' : 'true'); return; }
  });

  /* Klick auf den Backdrop schließt — erwartetes Verhalten, spart eine Regel */
  if (sheet) {
    sheet.addEventListener('click', function (e) { if (e.target === sheet) sheet.close(); });
  }

  /* ── Hell / Dunkel ────────────────────────────────────────────────
     Produktfunktion, kein Designer-Schalter. Der frühe Inline-Script im
     <head> setzt das Attribut vor dem ersten Paint, damit nichts blitzt. */
  document.addEventListener('click', function (e) {
    if (!e.target.closest('[data-theme-toggle]')) return;
    var root = document.documentElement;
    var next = root.dataset.theme === 'dark' ? 'light' : 'dark';
    root.dataset.theme = next;
    try { localStorage.setItem('lifelog-theme', next); } catch (err) {}
  });

  /* ── Finder: Auswahl in Baum und Quellenliste ─────────────────────
     htmx: <button class="tree-item" hx-get="/notizen?src=…"
                   hx-target=".finder-main" hx-swap="innerHTML"> */
  document.querySelectorAll('[data-finder-tree]').forEach(function (tree) {
    tree.addEventListener('click', function (e) {
      var item = e.target.closest('.tree-item');
      if (!item) return;
      tree.querySelectorAll('.tree-item').forEach(function (i) { i.removeAttribute('aria-current'); });
      item.setAttribute('aria-current', 'true');
    });
  });

  /* ── Git: Jahr umschalten ─────────────────────────────────────────
     htmx: <button hx-get="/git?jahr=2025" hx-target="#git-jahr"
                   hx-swap="outerHTML" hx-push-url="true">
     Dann liegt nur noch ein Jahr im Markup statt beider. */
  (function () {
    var sw = document.querySelector('[data-year-switch]');
    if (!sw) return;
    var blocks = Array.prototype.slice.call(document.querySelectorAll('.year-block'));
    var years  = blocks.map(function (b) { return parseInt(b.dataset.year, 10); }).sort();
    var label  = document.querySelector('[data-year-now]');
    var summary = document.querySelector('[data-year-label]');
    var TEXT = {};
    blocks.forEach(function (b) {
      var y = b.dataset.year;
      var big = b.querySelector('.gh-big').firstChild.nodeValue.trim();
      var repos = b.querySelector('.pane-label + .meta') ? '' : '';
      TEXT[y] = big;
    });
    var cur = years[years.length - 1];

    function show(y) {
      if (years.indexOf(y) === -1) return;
      cur = y;
      blocks.forEach(function (b) { b.hidden = parseInt(b.dataset.year, 10) !== y; });
      label.textContent = y;
      var b = blocks.filter(function (x) { return parseInt(x.dataset.year, 10) === y; })[0];
      var repos = b.querySelector('[data-od-id^="pane-listen"] .meta').textContent;
      var prs   = b.querySelectorAll('[data-od-id^="pane-listen"] .meta')[1].textContent;
      summary.textContent = TEXT[y] + ' Beiträge · ' + repos.replace(/^Top \d+ von /, '') +
                            ' Repos · ' + prs.replace(' im Jahr', '') + ' Pull Requests';
      sw.querySelectorAll('button').forEach(function (btn) {
        var next = y + parseInt(btn.dataset.step, 10);
        btn.disabled = years.indexOf(next) === -1;   // Aussehen kommt aus :disabled im CSS
      });
    }

    sw.addEventListener('click', function (e) {
      var btn = e.target.closest('button[data-step]');
      if (!btn || btn.disabled) return;
      show(cur + parseInt(btn.dataset.step, 10));
    });
    show(cur);
  })();
})();
