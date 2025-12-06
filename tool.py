import json
from PIL import Image, ImageSequence
import cv2
import numpy as np

output_file = "ciallo.json"
input_file = "out.gif"
source_gif_img = Image.open(input_file)

chars = [' ', '.', ',', ':', ';', '+', '*', '?', '%', '$', '#', '@', '@', '#', '$', '%']


def rgb_to_ansi_16(r, g, b, background=False):
    gray = 0.299 * r + 0.587 * g + 0.114 * b

    if gray < 32:
        color = 0
    elif gray < 64:
        color = 8
    elif gray < 96:
        color = 7
    elif gray < 128:
        color = 15
    elif gray < 160:
        color = 3
    elif gray < 192:
        color = 2
    elif gray < 224:
        color = 6
    else:
        color = 1

    if background:
        return f"\033[{color + 40}m"
    else:
        return f"\033[{color + 30}m"


def rgb_to_ansi_8bit(r, g, b, background=False):
    r6 = round(r / 255 * 5)
    g6 = round(g / 255 * 5)
    b6 = round(b / 255 * 5)

    color_216 = 16 + 36 * r6 + 6 * g6 + b6

    gray = (r + g + b) / 3
    gray_index = round(gray / 255 * 23)
    color_gray = 232 + gray_index

    def distance(r1, g1, b1, r2, g2, b2):
        return (r1 - r2) ** 2 + (g1 - g2) ** 2 + (b1 - b2) ** 2

    r216 = round(r6 * 255 / 5)
    g216 = round(g6 * 255 / 5)
    b216 = round(b6 * 255 / 5)

    gray_val = round(gray_index * 255 / 23)
    rgray = ggray = bgray = gray_val

    if distance(r, g, b, r216, g216, b216) < distance(r, g, b, rgray, ggray, bgray):
        color_index = color_216
    else:
        color_index = color_gray

    if background:
        return f"\033[48;5;{color_index}m"
    else:
        return f"\033[38;5;{color_index}m"


def rgb_to_ansi_24bit(r, g, b, background=False):
    if background:
        return f"\033[48;2;{r};{g};{b}m"
    else:
        return f"\033[38;2;{r};{g};{b}m"


def rgb_to_gray(r, g, b):
    return 0.299 * r + 0.587 * g + 0.114 * b


def process_gif(color_mode="24bit"):
    frames = []

    if color_mode == "gray":
        color_func = None
    elif color_mode == "16color":
        color_func = rgb_to_ansi_16
    elif color_mode == "8bit":
        color_func = rgb_to_ansi_8bit
    elif color_mode == "24bit":
        color_func = rgb_to_ansi_24bit
    else:
        raise ValueError("Invalid color mode. Choose from: gray, 16color, 8bit, 24bit")

    for frame in ImageSequence.Iterator(source_gif_img):
        frame_rgb = frame.convert("RGB")
        img = cv2.cvtColor(np.array(frame_rgb), cv2.COLOR_RGB2BGR)
        bilateral_blur = cv2.bilateralFilter(img, 9, 75, 75)
        frame_rgb = Image.fromarray(cv2.cvtColor(bilateral_blur, cv2.COLOR_BGR2RGB))
        ascii_frame = []
        for y in range(frame_rgb.height):
            line = ""
            for x in range(frame_rgb.width):
                r, g, b = frame_rgb.getpixel((x, y))
                gray = 0.299 * r + 0.587 * g + 0.114 * b
                c = chars[int(gray) // 16]

                if color_mode == "gray":
                    line += c
                else:
                    colored_char = f"{color_func(r, g, b)}{c}\033[0m"
                    line += colored_char
            ascii_frame.append(line)
        frames.append(ascii_frame)

    with open(output_file, "w") as f:
        f.write(json.dumps(frames, indent=4))
    print(f"Done, save to {output_file}")

# process_gif(color_mode="gray")
# process_gif(color_mode="16color")
# process_gif(color_mode="8bit")
process_gif(color_mode="24bit")
