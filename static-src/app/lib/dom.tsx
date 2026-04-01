interface DOMNode extends HTMLElement {
  dataset: DOMStringMap;
  attributes: NamedNodeMap;
  nodeName: string;
  parentNode: (Node & ParentNode) | null;
}

function isNode(target: EventTarget | null): target is Node {
  return target !== null && 'parentNode' in target;
}

function isHTMLElement(node: Node): node is HTMLElement {
  return 'nodeName' in node && 'dataset' in node;
}

export class DOM {
  static eventFilter(e: React.MouseEvent | MouseEvent, nodeName: string): DOMNode | null {
    const target = e.target;
    if (!isNode(target)) return null;

    let node: Node | null = target;
    while (node !== null) {
      if (isHTMLElement(node) && node.nodeName === nodeName) {
        // Safe: we've verified it's an HTMLElement with dataset
        return node as DOMNode;
      }
      node = node.parentNode;
    }
    return null;
  }
}