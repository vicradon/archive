# Largest Rectangle in Histogram

**LeetCode:** https://leetcode.com/problems/largest-rectangle-in-histogram/ (LC 84, Hard)

## Description

Given an array of bar heights (each bar has width 1), find the area of the largest rectangle that can fit within the histogram.

**Example:** `[2,1,5,6,2,3]` → `10`

## Idea

Another monotonic stack problem, one level harder than Daily Temperatures. Keep a stack of indices with **increasing** heights from bottom to top. For each bar (plus a sentinel `0`-height bar appended at the end, index `n`):

- While the current height is **less than** the height at the top of the stack, pop it — that popped bar can't extend any further right, so this is the moment to finalize its rectangle.
- The rectangle's height is `heights[popped]`.
- The rectangle's width is `i - stack.top() - 1` (or `i` if the stack is empty after popping), because `stack.top()` (after popping) is the nearest shorter bar to the left, and `i` is the nearest shorter bar to the right.
- Push the current index.

### Why the sentinel `i <= n` trick matters

Without it, any bars still on the stack at the end (e.g. a strictly increasing histogram like `[1,2,3]`) would never get popped, and you'd miss their rectangles. Pretending there's a "height 0" bar just past the end forces everything to resolve.

## Walkthrough on `[2,1,5,6,2,3]`

```
i=0 (h=2) → stack empty → push 0                    stack: [0]
i=1 (h=1) → 1<2(idx0) → pop 0
  → width = i=1 (stack empty)         area = 2*1 = 2 → maxArea=2
  → push 1                             stack: [1]
i=2 (h=5) → 5>1(idx1), no pop → push 2               stack: [1,2]
i=3 (h=6) → 6>5(idx2), no pop → push 3               stack: [1,2,3]
i=4 (h=2) → 2<6(idx3) → pop 3
  → width = 4-2-1=1   area=6*1=6 → maxArea=6
  → 2<5(idx2) → pop 2
  → width = 4-1-1=2   area=5*2=10 → maxArea=10
  → 2>1(idx1), no pop → push 4         stack: [1,4]
i=5 (h=3) → 3>2(idx4), no pop → push 5               stack: [1,4,5]
i=6 (sentinel h=0) → pop 5: width=6-4-1=1  area=3*1=3
  → pop 4: width=6-1-1=4  area=2*4=8
  → pop 1: width=6 (stack empty)  area=1*6=6
Final: maxArea = 10 — the rectangle of height 5 spanning indices 2-3 (bars 5,6), width 2.
```

## Complexity

O(n) time — same amortized argument as Daily Temperatures, each index pushed and popped at most once. O(n) space for the stack.

## Edge cases

- Strictly increasing: `[1,2,3,4]`
- Strictly decreasing: `[4,3,2,1]`
- All same height: `[3,3,3]` → `9` (full width, since nothing ever pops early)
- Single bar: `[5]` → `5`
