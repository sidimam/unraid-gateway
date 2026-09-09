#!/usr/bin/env python3
"""unraid-gateway icon from the author's artwork (Design/icon-source.png: Unraid bars, rounded
connector fading into the slate gateway arch). Writes:
  assets/icon.png        light version, 1024 px (Unraid template / Community Applications)
  assets/icon-dark.png   dark version derived from the same artwork (dark background, arch lightened)
  internal/webui/static/icon.png  256 px dark version used by the web UI (dark theme)
Pillow only."""
import os
from PIL import Image, ImageChops

HERE = os.path.dirname(__file__)
SRC = os.path.join(HERE, "..", "Design", "icon-source.png")
ASSETS = os.path.join(HERE, "..", "assets")
WEBUI = os.path.join(HERE, "..", "internal", "webui", "static")
S = 1024

def square(im, margin=0.06):
    """Pad (never crop) the artwork onto a square canvas of its own background colour."""
    w, h = im.size
    bg = im.crop((2, 2, 10, 10)).resize((1, 1), Image.BOX).getpixel((0, 0))
    side = int(max(w, h) * (1 + 2 * margin))
    canvas = Image.new("RGB", (side, side), bg)
    canvas.paste(im, ((side - w) // 2, (side - h) // 2))
    return canvas.resize((S, S), Image.LANCZOS)

def bars_mask(rgb):
    h, s, v = rgb.convert("HSV").split()
    warm = h.point(lambda x: 255 if (x < 50 or x > 241) else 0)
    sat = s.point(lambda x: 0 if x < 64 else min(255, (x - 64) * 5))
    bright = v.point(lambda x: 255 if x > 64 else 0)
    return ImageChops.multiply(ImageChops.multiply(warm, sat), bright)

def artwork_mask(rgb, lo=14, gain=6):
    """Alpha of the artwork against its own flat background. The dark variant uses a higher
    threshold so the soft drop shadow of the light artwork is not carried over."""
    bg = rgb.crop((2, 2, 10, 10)).resize((1, 1), Image.BOX).getpixel((0, 0))
    r, g, b = ImageChops.difference(rgb, Image.new("RGB", rgb.size, bg)).split()
    diff = ImageChops.lighter(ImageChops.lighter(r, g), b)  # max channel difference (keeps pale yellow)
    return diff.point(lambda x: 0 if x < lo else min(255, (x - lo) * gain))

def darken(light):
    """Dark background; the slate arch becomes light grey, bars untouched. Edge pixels are
    un-premultiplied against the light background so no pale halo remains on the dark one."""
    art = artwork_mask(light, lo=30, gain=6); bars = bars_mask(light)
    slate = ImageChops.subtract(art, bars)
    h, s, v = light.convert("HSV").split()
    v_inv = v.point(lambda x: min(255, 255 - x + 95))
    light_arch = Image.merge("HSV", (h, s.point(lambda x: x // 3), v_inv)).convert("RGB")
    fg = Image.composite(light_arch, light, slate)
    w, hgt = light.size
    bgc = light.crop((2, 2, 10, 10)).resize((1, 1), Image.BOX).getpixel((0, 0))
    out = Image.new("RGB", (w, hgt)); po = out.load(); pf = fg.load(); pa = art.load(); pl = light.load()
    top, bottom = (0x30, 0x35, 0x3D), (0x19, 0x1C, 0x22)
    for y in range(hgt):
        t = y / (hgt - 1); dbg = tuple(top[i] + (bottom[i] - top[i]) * t for i in range(3))
        for x in range(w):
            a = pa[x, y] / 255.0
            if a <= 0.0:
                po[x, y] = tuple(int(c) for c in dbg); continue
            src = pf[x, y] if a >= 0.999 else tuple(
                max(0, min(255, (pf[x, y][i] - (1 - a) * bgc[i]) / a)) for i in range(3))
            po[x, y] = tuple(int(src[i] * a + dbg[i] * (1 - a)) for i in range(3))
    return out

def main():
    light = square(Image.open(SRC).convert("RGB"))
    dark = darken(light)
    os.makedirs(ASSETS, exist_ok=True)
    light.save(os.path.join(ASSETS, "icon.png"))
    dark.save(os.path.join(ASSETS, "icon-dark.png"))
    dark.resize((256, 256), Image.LANCZOS).save(os.path.join(WEBUI, "icon.png"))
    print("written", os.path.abspath(ASSETS), "and web UI icon")

if __name__ == "__main__":
    main()
