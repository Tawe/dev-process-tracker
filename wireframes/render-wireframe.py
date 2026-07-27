#!/usr/bin/env python3
"""Render tui-wireframe markup as a colored terminal HTML page.

Usage:
    python3 render-wireframe.py wireframe.txt
    python3 render-wireframe.py wireframe.txt -o output.html
    python3 render-wireframe.py wireframe.txt --no-open

Reads wireframe blocks from a markdown file (```wireframe ... ```),
converts {tag}text{/} markup to colored HTML spans, wraps in a
terminal frame, and opens in the default browser.
"""

import re
import sys
import os
import webbrowser
import argparse
import tempfile

TAG_MAP = {
    'g': 'g', 'r': 'r', 'y': 'y', 'b': 'b', 'c': 'c', 'p': 'p',
    'w': 'w', 'gr': 'gr', 'm': 'm', 'o': 'o',
    'bold': 'bold', 'dim': 'dim', 'sel': 'sel', 'inv': 'inv',
}

TAG_NAMES = sorted(TAG_MAP.keys(), key=len, reverse=True)
OPEN_TAG_PATTERN = re.compile(r'\{(' + '|'.join(re.escape(t) for t in TAG_NAMES) + r')\}')


def visible_length(line):
    """Length of a line with all markup tags stripped."""
    s = re.sub(r'\{(?:g|r|y|b|c|p|w|gr|m|o|bold|dim|sel|inv)\}', '', line)
    s = re.sub(r'\{/\}', '', s)
    return len(s)


def parse_markup(text):
    """Convert {tag}text{/} markup to HTML spans using a real stack parser."""
    def esc(ch):
        return ch.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;')

    out = []
    stack = []
    i = 0
    while i < len(text):
        if text.startswith('{/}', i):
            if stack:
                stack.pop()
                out.append('</span>')
            else:
                out.append('{/}')
            i += 3
            continue

        match = OPEN_TAG_PATTERN.match(text, i)
        if match:
            tag = match.group(1)
            stack.append(tag)
            out.append(f'<span class="{TAG_MAP[tag]}">')
            i = match.end()
            continue

        out.append(esc(text[i]))
        i += 1

    # Close unclosed spans so malformed wireframes do not break the whole page.
    while stack:
        stack.pop()
        out.append('</span>')

    return ''.join(out)


def validate_widths(text, max_width=120):
    """Check visible line widths. Returns list of (line_num, width, issue)."""
    issues = []
    for i, line in enumerate(text.split('\n'), 1):
        w = visible_length(line)
        if w > max_width:
            issues.append((i, w, f'OVER by {w - max_width}'))
        elif w >= max_width - 2:
            issues.append((i, w, f'near limit'))
    return issues


CSS = """\
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { background: #1a1b26; font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Menlo', monospace; font-size: 12.5px; line-height: 1.5; padding: 16px; }
  .frame { border-radius: 10px; overflow: hidden; box-shadow: 0 4px 24px rgba(0,0,0,0.4); margin-bottom: 20px; }
  .bar { background: #16161e; padding: 7px 14px; display: flex; align-items: center; gap: 7px; border-bottom: 1px solid #292e42; }
  .dot { width: 10px; height: 10px; border-radius: 50%; }
  .dr { background: #ff5f57; } .dy { background: #febc2e; } .dg { background: #28c840; }
  .bar-title { flex:1; text-align:center; color:#565f89; font-size:11px; }
  .body { padding: 10px 16px 8px; white-space: pre; color: #a9b1d6; overflow-x: auto; }
  .g { color: #9ece6a; } .r { color: #f7768e; } .y { color: #e0af68; } .b { color: #7aa2f7; }
  .c { color: #7dcfff; } .p { color: #bb9af7; } .w { color: #c0caf5; } .gr { color: #565f89; }
  .m { color: #3b4261; } .o { color: #ff9e64; } .bold { font-weight: 700; }
  .sel { background: #283457; } .dim { opacity: 0.55; }
  .inv { background: #a9b1d6; color: #1a1b26; padding: 0; border-radius: 0; }
  .state-label { text-align: center; color: #565f89; font-size: 11px; padding: 6px 0; letter-spacing: 0.5px; }
  .validation { color: #f7768e; font-family: sans-serif; font-size: 13px; padding: 8px 12px; margin-bottom: 12px; border-radius: 6px; background: #2a1a1a; }
  .validation.pass { color: #9ece6a; background: #1a2a1a; }
"""


def render_frame(title, state_label, content_html):
    """Wrap parsed content in a terminal frame."""
    label = f'<div class="state-label">{state_label}</div>\n' if state_label else ''
    return f"""{label}<div class="frame">
  <div class="bar">
    <div class="dot dr"></div><div class="dot dy"></div><div class="dot dg"></div>
    <span class="bar-title">{title}</span>
  </div>
  <div class="body">{content_html}</div>
</div>
"""


def extract_wireframes(content):
    """Extract wireframe blocks from markdown content.

    Returns list of (state_label, viewport_width, markup_text).
    """
    blocks = re.findall(r'```wireframe\s*(.*?)\n(.*?)```', content, re.DOTALL)
    results = []
    for header, text in blocks:
        state_match = re.search(r'state:(\S+)', header)
        vp_match = re.search(r'viewport:(\d+)', header)
        state = state_match.group(1) if state_match else 'default'
        viewport = int(vp_match.group(1)) if vp_match else 120
        results.append((state, viewport, text.rstrip()))
    return results


def main():
    parser = argparse.ArgumentParser(description='Render tui-wireframe markup as colored HTML')
    parser.add_argument('input', help='Input file with wireframe blocks')
    parser.add_argument('-o', '--output', help='Output HTML file (default: temp file)')
    parser.add_argument('--no-open', action='store_true', help="Don't open in browser")
    parser.add_argument('--max-width', type=int, default=120, help='Max width for validation (default: 120)')
    parser.add_argument('--title', default='devpt', help='Title bar text')
    args = parser.parse_args()

    with open(args.input) as f:
        content = f.read()

    wireframes = extract_wireframes(content)

    if not wireframes:
        print("No ```wireframe blocks found in input.", file=sys.stderr)
        sys.exit(1)

    frames_html = []
    all_pass = True

    for state, viewport, markup in wireframes:
        issues = validate_widths(markup, max_width=args.max_width)
        if issues:
            all_pass = False

        parsed = parse_markup(markup)
        label = f'state: {state}  viewport: {viewport}'
        frames_html.append(render_frame(args.title, label, parsed))

    validation_html = ''
    if all_pass:
        validation_html = f'<div class="validation pass">All lines within {args.max_width} chars</div>'
    else:
        lines = [f'<div class="validation">Width validation issues:</div>']
        for state, viewport, markup in wireframes:
            issues = validate_widths(markup, max_width=args.max_width)
            for line_num, width, issue in issues:
                lines.append(f'<div class="validation">  {state} line {line_num}: {width} chars ({issue})</div>')
        validation_html = '\n'.join(lines)

    full_html = f"""<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Wireframe Preview</title>
<style>{CSS}</style>
</head>
<body>
{validation_html}
{"".join(frames_html)}
</body>
</html>"""

    if args.output:
        out_path = args.output
    else:
        tmp = tempfile.NamedTemporaryFile(suffix='.html', delete=False, prefix='wireframe-')
        out_path = tmp.name
        tmp.close()

    with open(out_path, 'w') as f:
        f.write(full_html)

    print(f"Rendered {len(wireframes)} state(s) to {out_path}")

    if not args.no_open:
        webbrowser.open(f'file://{os.path.abspath(out_path)}')


if __name__ == '__main__':
    main()
