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

def square(im):
    w, h = im.size; side = min(w, h)
    return im.crop(((w - side) // 2, (h - side) // 2, (w + side) // 2, (h + side) // 2)).resize((S, S), Image.LANCZOS)

def bars_mask(rgb):
    h, s, v = rgb.convert("HSV").split()
    warm = h.point(lambda x: 255 if (x < 50 or x > 241) else 0)
    sat = s.point(lambda x: 0 if x < 64 else min(255, (x - 64) * 5))
    bright = v.point(lambda x: 255 if x > 64 else 0)
    return ImageChops.multiply(ImageChops.multiply(warm, sat), bright)

def artwork_mask(rgb):
    bg = rgb.crop((2, 2, 10, 10)).resize((1, 1), Image.BOX).getpixel((0, 0))
    diff = ImageChops.difference(rgb, Image.new("RGB", rgb.size, bg)).convert("L")
    return diff.point(lambda x: 0 if x < 14 else min(255, (x - 14) * 6))

def darken(light):
    """Dark background; the slate arch becomes light grey (value inverted), bars untouched."""
    art = artwork_mask(light); bars = bars_mask(light)
    slate = ImageChops.subtract(art, bars)
    h, s, v = light.convert("HSV").split()
    v_inv = v.point(lambda x: min(255, 255 - x + 95))
    light_arch = Image.merge("HSV", (h, s.point(lambda x: x // 3), v_inv)).convert("RGB")
    with_arch = Image.composite(light_arch, light, slate)
    w, hgt = light.size
    bg = Image.new("RGB", (w, hgt)); px = bg.load()
    top, bottom = (0x30, 0x35, 0x3D), (0x19, 0x1C, 0x22)
    for y in range(hgt):
        t = y / (hgt - 1); c = tuple(int(top[i] + (bottom[i] - top[i]) * t) for i in range(3))
        for x in range(w): px[x, y] = c
    return Image.composite(with_arch, bg, art)

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
