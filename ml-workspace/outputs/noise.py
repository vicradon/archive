import numpy as np
from PIL import Image
import sys
import math

T = 1000
BETA_START = 1e-4
BETA_END = 0.02

def build_schedule():
    betas = np.linspace(BETA_START, BETA_END, T)
    alphas = 1 - betas
    return np.cumprod(alphas)

def noise_step(img: Image.Image, t: int, alpha_bar: np.ndarray) -> Image.Image:
    a = alpha_bar[t]
    arr = np.array(img).astype(np.float32) / 127.5 - 1.0
    noise = np.random.randn(*arr.shape).astype(np.float32)
    noised = math.sqrt(a) * arr + math.sqrt(1 - a) * noise
    noised = ((noised + 1) / 2 * 255).clip(0, 255).astype(np.uint8)
    return Image.fromarray(noised, mode=img.mode)

if __name__ == "__main__":
    import urllib.request

    src, t = sys.argv[1], int(sys.argv[2])
    out_path = sys.argv[3] if len(sys.argv) >= 4 else f"noised_{t}.png"

    if src.startswith("http://") or src.startswith("https://"):
        with urllib.request.urlopen(src) as resp:
            img = Image.open(resp).copy()
    else:
        img = Image.open(src)

    alpha_bar = build_schedule()
    noised = noise_step(img, t, alpha_bar)
    noised.save(out_path)
    print(f"saved -> {out_path}")
