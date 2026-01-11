// Vector Math Utilities
const VectorMath = {
    parse(input) {
        input = input.trim();
        if (input.startsWith('[[') && input.endsWith(']]')) {
            const components = input.slice(2, -2).split('],[');
            return components.map(c => parseFloat(c.trim()));
        } else if (input.startsWith('[') && input.endsWith(']')) {
            const components = input.slice(1, -1).split(',');
            return components.map(c => parseFloat(c.trim()));
        }
        return null;
    },

    dotProduct(v1, v2) {
        return v1.reduce((sum, val, i) => sum + val * v2[i], 0);
    },

    magnitude(v) {
        return Math.sqrt(v.reduce((sum, val) => sum + val * val, 0));
    },

    cosineSimilarity(v1, v2) {
        const dot = this.dotProduct(v1, v2);
        const mag1 = this.magnitude(v1);
        const mag2 = this.magnitude(v2);
        if (mag1 === 0 || mag2 === 0) return 0;
        return dot / (mag1 * mag2);
    },

    normalize(v) {
        const mag = this.magnitude(v);
        if (mag === 0) return v;
        return v.map(val => val / mag);
    }
};
