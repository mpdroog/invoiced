interface DOMNode extends HTMLElement {
  dataset: DOMStringMap;
  attributes: NamedNodeMap;
  nodeName: string;
  parentNode: (Node & ParentNode) | null;
}

export class DOM {
  static eventFilter(e: React.MouseEvent | MouseEvent, nodeName: string): DOMNode | null {
    if (e.target) {
      let node: Node | null = e.target as Node;
      while (node !== null) {
        if ((node as HTMLElement).nodeName === nodeName) {
          return node as DOMNode;
        }
        node = node.parentNode;
      }
    }
    return null;
  }
}