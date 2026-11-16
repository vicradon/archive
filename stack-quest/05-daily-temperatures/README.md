# Daily Temperatures

**LeetCode:** https://leetcode.com/problems/daily-temperatures/ (LC 739, Medium)

## Description

Given an array of daily temperatures, return an array where each element tells you how many days you'd have to wait until a warmer temperature. If there's no future day with a warmer temperature, put `0`.

**Example:** `[73,74,75,71,69,72,76,73]` → `[1,1,4,2,1,1,0,0]`

## Idea

This is the classic **monotonic stack** pattern. The naive approach is, for each day, scan forward until you find a warmer day — that's O(n²). We can do better by keeping a stack of "days that are still waiting for a warmer day," and only ever looking backward once per new temperature.

Walk through the temperatures left to right. Keep a stack of *indices* (not values), and maintain the invariant that the stack is always **decreasing** in temperature from bottom to top. For each new day:

1. While the stack isn't empty and the current temperature is **warmer** than the temperature at the index on top of the stack — that's the answer for that index! Pop it, and record `currentIndex - poppedIndex` as its wait time.
2. Push the current index onto the stack (it's now waiting for its own warmer day).

Any index still sitting on the stack at the end never found a warmer day, so those stay `0`.

## Complexity

O(n) time — each index is pushed and popped at most once (amortized). O(n) space for the stack.
