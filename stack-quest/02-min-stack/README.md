# Min Stack

**LeetCode:** https://leetcode.com/problems/min-stack/ (LC 155, Medium)

## Description

Design a stack that supports `push`, `pop`, `top`, and `getMin` — all in O(1) time.

## The challenge

A normal stack gives you O(1) push/pop/top, but if you ask "what's the minimum right now?" you'd normally have to scan the whole stack — O(n). We want O(1) for that too.

## Idea

Keep two stacks in parallel:

- A regular stack holding all the values.
- A "min stack" where each position tracks the minimum value *at that point in the stack's history*.

Every time you push a value, you also push the current minimum (either the new value, if it's smaller, or the previous minimum) onto the min stack. When you pop, you pop from both stacks together — so the min stack always reflects the correct minimum for whatever's currently in the main stack.

## Complexity

O(1) time for all operations, O(n) space (two parallel stacks).
