#!/usr/bin/env python3
"""unraid-gateway icon: the Unraid bars flowing through an arrow into a gateway arch that
holds a folder and a network port. Light version (white → pale grey background) as requested.
Writes assets/icon.png (1024) and assets/icon.svg. Needs Pillow."""
import os
from PIL import Image, ImageDraw, ImageFilter

S, SS = 1024, 4
N = S * SS
OUT = os.path.join(os.path.dirname(__file__), "..", "assets")

def P(x, y): return (x * N, y * N)
def lerp(a, b, t): return tuple(int(a[i] + (b[i] - a[i]) * t) for i in range(3))

def vgrad(size, top, bottom, alpha=255):
    img = Image.new("RGBA", size); px = img.load()
    for y in range(size[1]):
        c = lerp(top, bottom, y / max(size[1] - 1, 1))
        for x in range(size[0]): px[x, y] = c + (alpha,)
    return img

ORANGE_TOP, ORANGE_BOT = (0xFF, 0xC2, 0x4A), (0xF0, 0x4E, 0x2A)   # yellow-orange → Unraid red-orange
GREY = (0x9A, 0xA1, 0xAA, 255)
GREY_LIGHT = (0xC9, 0xCE, 0xD4, 255)

def bar(draw_layer, x0, y0, x1, y1, slant, r):
    """Slanted bar (parallelogram with rounded ends) filled with the orange gradient."""
    w, h = int(x1 - x0), int(y1 - y0)
    g = vgrad((w, h), ORANGE_TOP, ORANGE_BOT)
    m = Image.new("L", (w, h), 0); d = ImageDraw.Draw(m)
    d.polygon([(0, slant), (w, 0), (w, h - slant), (0, h)], fill=255)
    m = m.filter(ImageFilter.GaussianBlur(1))
    draw_layer.paste(g, (int(x0), int(y0)), m)

def build():
    # background: white → pale grey with a soft vignette
    bg = vgrad((N, N), (0xFF, 0xFF, 0xFF), (0xE6, 0xE9, 0xEE))
    L = Image.new("RGBA", (N, N), (0, 0, 0, 0))
    d = ImageDraw.Draw(L)
    # three Unraid bars (like the logo mark), heights and slants
    bars = [(0.14, 0.32, 0.24, 0.72), (0.29, 0.22, 0.37, 0.76), (0.42, 0.34, 0.49, 0.66)]
    for i, (x0, y0, x1, y1) in enumerate(bars):
        bar(L, x0 * N, y0 * N, x1 * N, y1 * N, slant=int(0.045 * N) if i == 0 else int(0.02 * N), r=0)
    # arrow from the bars to the arch
    ax0, ay = 0.47 * N, 0.505 * N
    stroke = int(0.045 * N)
    d.line([(ax0, ay), (0.70 * N, ay)], fill=(0xF6, 0x8A, 0x3A, 255), width=stroke)
    d.polygon([(0.68 * N, ay - 0.075 * N), (0.78 * N, ay), (0.68 * N, ay + 0.075 * N)], fill=(0xF6, 0x8A, 0x3A, 255))
    # gateway arch (door) outline in grey
    arch = Image.new("RGBA", (N, N), (0, 0, 0, 0)); ad = ImageDraw.Draw(arch)
    x0, x1, ytop, ybot = 0.55 * N, 0.87 * N, 0.26 * N, 0.74 * N
    aw = int(0.05 * N)
    ad.rounded_rectangle([x0, ytop, x1, ybot], radius=int((x1 - x0) / 2), outline=GREY, width=aw)
    ad.rectangle([x0, ybot - 0.12 * N, x1, ybot], fill=(0, 0, 0, 0))
    ad.rounded_rectangle([x0, ytop + 0.16 * N, x1, ybot], radius=int(0.03 * N), outline=GREY, width=aw)
    # folder glyph inside the arch
    fx0, fy0, fx1, fy1 = 0.655 * N, 0.36 * N, 0.765 * N, 0.44 * N
    ad.rounded_rectangle([fx0, fy0 + 0.012 * N, fx1, fy1], radius=int(0.01 * N), fill=GREY_LIGHT)
    ad.rounded_rectangle([fx0, fy0, fx0 + 0.045 * N, fy0 + 0.025 * N], radius=int(0.006 * N), fill=GREY_LIGHT)
    # network port glyph at the bottom of the arch
    px0, py0, px1, py1 = 0.675 * N, 0.62 * N, 0.745 * N, 0.68 * N
    ad.rounded_rectangle([px0, py0, px1, py1], radius=int(0.008 * N), outline=GREY_LIGHT, width=int(0.012 * N))
    for i in range(6):
        x = px0 + 0.012 * N + i * 0.0095 * N
        ad.line([(x, py0 + 0.012 * N), (x, py0 + 0.03 * N)], fill=GREY_LIGHT, width=int(0.004 * N))
    glow = arch.filter(ImageFilter.GaussianBlur(int(0.01 * N)))
    out = Image.new("RGBA", (N, N)); out.alpha_composite(bg); out.alpha_composite(glow); out.alpha_composite(arch); out.alpha_composite(L)
    return out.resize((S, S), Image.LANCZOS)

SVG = f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
<defs>
 <linearGradient id="bg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffffff"/><stop offset="1" stop-color="#e6e9ee"/></linearGradient>
 <linearGradient id="or" x1="0" y1="0" x2="0" y2="1" gradientUnits="objectBoundingBox"><stop offset="0" stop-color="#ffc24a"/><stop offset="1" stop-color="#f04e2a"/></linearGradient>
</defs>
<rect width="1024" height="1024" fill="url(#bg)"/>
<polygon points="143,373 246,328 246,691 143,737" fill="url(#or)"/>
<polygon points="297,246 379,225 379,758 297,778" fill="url(#or)"/>
<polygon points="430,368 502,348 502,656 430,676" fill="url(#or)"/>
<line x1="481" y1="517" x2="717" y2="517" stroke="#f68a3a" stroke-width="46" stroke-linecap="round"/>
<polygon points="696,440 799,517 696,594" fill="#f68a3a"/>
<path d="M563 758 V430 A164 164 0 0 1 891 430 V758" fill="none" stroke="#9aa1aa" stroke-width="51" stroke-linecap="round"/>
<rect x="563" y="430" width="328" height="328" rx="31" fill="none" stroke="#9aa1aa" stroke-width="51"/>
<rect x="671" y="381" width="112" height="70" rx="10" fill="#c9ced4"/><rect x="671" y="369" width="46" height="26" rx="6" fill="#c9ced4"/>
<rect x="691" y="635" width="72" height="61" rx="8" fill="none" stroke="#c9ced4" stroke-width="12"/>
</svg>
'''

if __name__ == "__main__":
    os.makedirs(OUT, exist_ok=True)
    img = build()
    img.convert("RGB").save(os.path.join(OUT, "icon.png"))
    open(os.path.join(OUT, "icon.svg"), "w").write(SVG)
    print("written", os.path.abspath(OUT))
