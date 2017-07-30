export class DOM {
  static eventFilter(e: any, nodeName: string): BrowserEventTarget {
    if (e.target) {
      var node = e.target;
      while (node !== window && node !== null) {
        if (node.nodeName === nodeName) {
          return node;
        }
        node = node.parentNode;
      }
    }
    return null;
  }
}