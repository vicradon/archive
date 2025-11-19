# Pi Calculator

A JavaScript implementation of the Chudnovsky algorithm for calculating π to high precision using decimal.js.

## Introduction

This project calculates π using the Chudnovsky algorithm, which rapidly converges to provide high-precision results. The implementation computes π to specified decimal places using arbitrary-precision arithmetic.

The Chudnovsky algorithm is based on the series:

$$\frac{1}{\pi} = 12 \sum_{q=0}^{\infty} \frac{(-1)^q (6q)! (545140134q + 13591409)}{(3q)! (q!)^3 (640320)^{3q + 3/2}}$$

## Clone and Install

```bash
git clone https://github.com/vicradon/pi-calculator
cd pi
npm install
```

## Run

```bash
node compute.js
```

The script will output π calculated to 7 and 1000 decimal places.

## Contributions

Contributions are welcome. Please ensure any changes maintain the mathematical accuracy of the Chudnovsky algorithm implementation.