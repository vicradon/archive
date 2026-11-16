# shapegen — 2D Shape Dataset Generator

Generate synthetic polygon image datasets for computer vision training.
Pure Go, zero dependencies, single binary.

## Classes

| Name | Sides |
|------|-------|
| triangle | 3 |
| quadrilateral | 4 |
| pentagon | 5 |
| hexagon | 6 |
| heptagon | 7 |
| octagon | 8 |
| decagon | 10 |

## Usage

```sh
# Default: all 7 classes, 100 images each, 40×40px, COCO format
./shapegen

# Custom size and count
./shapegen -size 64 -n 200

# Specific classes only
./shapegen -classes triangle,hexagon,decagon -n 50

# YOLO format
./shapegen -format yolo -o yolo_dataset.zip

# Large dataset with more workers
./shapegen -size 128 -n 500 -workers 8 -format coco

# Fixed seed for reproducibility
./shapegen -seed 42 -n 100
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-size` | `40` | Image size in pixels (square) |
| `-n` | `100` | Images per class |
| `-format` | `coco` | Export format: `plain`, `coco`, `yolo`, `voc` |
| `-o` | auto | Output ZIP filename |
| `-classes` | `all` | Comma-separated class names or `all` |
| `-workers` | `4` | Parallel goroutines |
| `-seed` | `0` | Random seed (0 = time-based) |
| `-list` | — | List all classes and exit |

## Export Formats

- **plain** — Images organized into folders by class name
- **coco** — COCO-format JSON (`annotations/instances.json`) + `images/`
- **yolo** — YOLO `.txt` labels + `data.yaml` config
- **voc** — Pascal VOC XML annotations per image in `Annotations/`

## Building from source

```sh
go build -o shapegen .
```
