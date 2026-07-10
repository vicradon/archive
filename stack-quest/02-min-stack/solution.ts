class MinStack {
  private stack: number[] = [];
  private minStack: number[] = [];

  private getLast(arr: number[]): number {
    if (arr.length === 0) {
      throw new Error("Cannot get last element from empty array");
    }
    const lastIndex = arr.length - 1;
    const value = arr[lastIndex];
    if (value === undefined) {
      throw new Error("Unexpected undefined value");
    }
    return value;
  }

  push(val: number): void {
    this.stack.push(val);
    if (this.minStack.length === 0) {
      this.minStack.push(val);
    } else {
      const previousMin = this.getLast(this.minStack);
      const currentMin = Math.min(val, previousMin);
      this.minStack.push(currentMin);
    }
  }

  pop(): void {
    if (this.stack.length === 0) {
      throw new Error("Cannot pop from an empty stack");
    }
    this.stack.pop();
    this.minStack.pop();
  }

  top(): number {
    return this.getLast(this.stack);
  }

  getMin(): number {
    return this.getLast(this.minStack);
  }
}


const minStack = new MinStack()

minStack.push(6)
minStack.push(5)
minStack.push(9)
minStack.push(1)
minStack.push(20)
minStack.push(11)
console.log(minStack.getMin())
minStack.pop()
minStack.pop()
minStack.pop()
console.log(minStack.getMin())
