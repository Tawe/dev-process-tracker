# Wireframes

Terminal UI mockups for devpt, authored as text and rendered to a colored
HTML page that mirrors the real TUI.

## Run

```bash
python3 wireframes/render-wireframe.py wireframes/wireframe.md
```

Opens a self-contained HTML page in the browser. No server, no port — the
script reads the markup, writes one HTML file, and opens it via `file://`.
Add `--no-open` to skip the browser, `-o out.html` to pick the output path.

## How it works

`wireframe.md` holds one or more fenced blocks:

    ```wireframe state:edit-modal viewport:120
    ...terminal art with {tag}markup{/}...
    ```

- `state:` / `viewport:` are labels shown above each frame.
- `{tag}...{/}` wraps text in a color/style. Tags nest; `{/}` closes the
  most recent open tag.

### Tag vocabulary

| tag   | style                | typical use                 |
|-------|----------------------|-----------------------------|
| `{w}` | white, bold          | app title                   |
| `{b}` | blue                 | headers, column titles      |
| `{c}` | cyan                 | service names               |
| `{g}` | green                | running / healthy           |
| `{y}` | yellow               | warning / sort indicator    |
| `{r}` | red                  | crashed / error             |
| `{o}` | orange               | accent                      |
| `{gr}`| gray                 | labels, dim metadata        |
| `{m}` | faint                | separators                  |
| `{sel}`| selection background | cursor row / field          |
| `{inv}`| inverse keycap       | key hints (`{inv}tab{/}`)   |
| `{bold}` / `{dim}` | weight / opacity | emphasis              |

Palette mirrors the Tokyo Night theme used by the TUI mockups; the live Go
TUI emits ANSI codes (`8 2 10 11 12 208`) resolved by the terminal theme.

## Width check

The renderer validates every line against `viewport` (default 120) and flags
over-limit lines. Author art at ≤120 display columns. Note: validation counts
codepoints, so wide/CJK glyphs may under-report — keep glyphs single-width.

## Files

- `wireframe.md` — active mockups (`default`, `filter-active`, `error`,
  `edit-modal`, `add-modal`). The `*-modal` states cover DEVPT-020 (add/edit
  service form; `edit-modal` shows rename via the editable Name field).
- `render-wireframe.py` — the renderer.
- `wireframe.txt`, `mockup.txt` — archived earlier drafts (raw terminal
  captures from v0.5.0 / v0.4.1). Not part of the active pipeline.
