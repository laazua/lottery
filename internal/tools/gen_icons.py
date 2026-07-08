#!/usr/bin/env python3
"""从 SVG 源文件生成 Android 各分辨率 PNG 图标。"""

import cairosvg
import os
import sys


def main():
    svg_path = sys.argv[1] if len(sys.argv) > 1 else "ui/icon/app_icon.svg"
    out_dir = sys.argv[2] if len(sys.argv) > 2 else "android/ic_launcher"

    sizes = {
        "mipmap-mdpi": 48,
        "mipmap-hdpi": 72,
        "mipmap-xhdpi": 96,
        "mipmap-xxhdpi": 144,
        "mipmap-xxxhdpi": 192,
    }

    for density, size in sizes.items():
        d = os.path.join(out_dir, density)
        os.makedirs(d, exist_ok=True)
        output = os.path.join(d, "ic_launcher.png")
        cairosvg.svg2png(url=svg_path, write_to=output, output_width=size, output_height=size)
        print(f"  {density}: {size}x{size} -> {output}")

    # XXX: 在根目录存一份，供 gogio -icon 使用
    root_png = os.path.join(out_dir, "ic_launcher.png")
    cairosvg.svg2png(url=svg_path, write_to=root_png, output_width=192, output_height=192)
    print(f"  root: 192x192 -> {root_png}")

    print("✅ 图标生成完成")


if __name__ == "__main__":
    main()
