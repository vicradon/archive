import Decimal from 'decimal.js';
/**
 * Calculate π using the Chudnovsky algorithm.
 *
 * The Chudnovsky algorithm is based on the series:
 *
 * $$\frac{1}{\pi} = 12 \sum_{q=0}^{\infty} \frac{(-1)^q (6q)! (545140134q + 13591409)}{(3q)! (q!)^3 (640320)^{3q + 3/2}}$$
 *
 * @param {number} digits - Number of decimal places to compute
 * @returns {string} π calculated to the specified precision
 */
function calcPi(digits) {
    const terms = Math.floor(digits / 13) + 1;
    Decimal.set({ precision: digits + 10 });

    let result = new Decimal(0);

    for (let q = 0; q < terms; q++) {
        const sign = q % 2 === 0 ? 1 : -1;
        const n = new Decimal(sign).mul(factorial(6 * q)).mul(545140134 * q + 13591409);
        const dev = factorial(3 * q).mul(factorial(q).pow(3)).mul(new Decimal(640320).pow(3 * q));
        result = result.add(n.div(dev));
    }

    result = result.mul(12).div(new Decimal(640320).pow(1.5));
    return new Decimal(1).div(result).toFixed(digits);
}

function factorial(n) {
    let result = new Decimal(1);
    for (let i = 2; i <= n; i++) {
        result = result.mul(i);
    }
    return result;
}

console.log(calcPi(7));
console.log(calcPi(1000));
