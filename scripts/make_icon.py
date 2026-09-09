#!/usr/bin/env python3
"""unraid-gateway icon, coherent with the Unraid Drive app icon: the same Unraid bars artwork
(Design/bars-source.png, shared with the app) feeding a short connector into the gateway arch,
drawn in the same slate grey with the same soft shadow. Writes assets/icon.png (1024). Pillow only."""
import os
from PIL import Image, ImageChops, ImageDraw, ImageFilter

HERE = os.path.dirname(__file__)
SRC = os.path.join(HERE, "..", "Design", "bars-source.png")
OUT = os.path.join(HERE, "..", "assets", "icon.png")
S = 1024

def bars_mask(rgb):
    h, s, v = rgb.convert("HSV").split()
    warm = h.point(lambda x: 255 if (x < 50 or x > 241) else 0)
    sat = s.point(lambda x: 0 if x < 64 else min(255, (x - 64) * 5))
    bright = v.point(lambda x: 255 if x > 64 else 0)
    return ImageChops.multiply(ImageChops.multiply(warm, sat), bright)

def main():
    src = Image.open(SRC).convert("RGB")
    w, h = src.size; side = min(w, h)
    src = src.crop(((w - side) // 2, (h - side) // 2, (w + side) // 2, (h + side) // 2)).resize((S, S), Image.LANCZOS)
    bg_col = src.crop((2, 2, 10, 10)).resize((1, 1), Image.BOX).getpixel((0, 0))
    bars = bars_mask(src)
    # keep only the three bars: the artwork's connector line starts right of the third bar,
    # so measure the bars' extent on the upper part of the image (above the line) and cut there
    upper = bars.crop((0, 0, S, int(S * 0.45)))
    bx0, _, bx1, _ = upper.getbbox()
    cut = Image.new("L", (S, S), 0); ImageDraw.Draw(cut).rectangle([0, 0, bx1 + 2, S], fill=255)
    bars = ImageChops.multiply(bars, cut)
    bars_soft = bars.filter(ImageFilter.GaussianBlur(1))
    canvas = Image.new("RGB", (S, S), bg_col)
    # soft shadow like the artwork
    shadow = Image.new("RGBA", (S, S), (0, 0, 0, 0)); shadow.paste((0, 0, 0, 70), (6, 8), bars)
    shadow = shadow.filter(ImageFilter.GaussianBlur(10))
    # slate arch: thick rounded "∩" with legs, plus a short stub from the third bar
    slate_top, slate_bot = (0x55, 0x61, 0x6D), (0x3A, 0x43, 0x4D)
    SS = 4; N = S * SS
    arch = Image.new("L", (N, N), 0); d = ImageDraw.Draw(arch)
    stroke = int(0.078 * N)
    ax0, ax1 = int((bx1 / S + 0.055) * N), int(0.945 * N)
    atop, abot = int(0.24 * N), int(0.79 * N)
    r = int((ax1 - ax0) * 0.42)
    d.rounded_rectangle([ax0, atop, ax1, abot + r], radius=r, outline=255, width=stroke)
    d.rectangle([0, abot, N, N], fill=0)                     # open the bottom: legs end flat
    for x in (ax0, ax1 - stroke):                            # round the leg ends
        d.rectangle([x, abot - stroke // 2, x + stroke, abot], fill=255)
        d.ellipse([x, abot - stroke // 2, x + stroke, abot + stroke // 2], fill=255)
    d.rectangle([0, abot + stroke // 2 + 1, N, N], fill=0)
    # stub connector from the bars to the left leg, mid height
    sy = int(0.56 * N); st = int(0.052 * N)
    d.rounded_rectangle([int((bx1 / S - 0.01) * N), sy - st // 2, ax0 + stroke // 2, sy + st // 2], radius=st // 3, fill=255)
    arch = arch.resize((S, S), Image.LANCZOS)
    grad = Image.new("RGB", (S, S)); px = grad.load()
    for y in range(S):
        t = y / (S - 1); c = tuple(int(slate_top[i] + (slate_bot[i] - slate_top[i]) * t) for i in range(3))
        for x in range(S): px[x, y] = c
    arch_shadow = Image.new("RGBA", (S, S), (0, 0, 0, 0)); arch_shadow.paste((0, 0, 0, 80), (6, 8), arch)
    arch_shadow = arch_shadow.filter(ImageFilter.GaussianBlur(10))
    out = canvas.convert("RGBA")
    out.alpha_composite(shadow); out.alpha_composite(arch_shadow)
    out = Image.composite(grad.convert("RGBA"), out, arch)
    out = Image.composite(src.convert("RGBA"), out, bars_soft)
    out.convert("RGB").save(OUT)
    print("written", os.path.abspath(OUT))

if __name__ == "__main__":
    main()
