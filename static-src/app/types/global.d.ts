// Global type declarations for the invoiced application

// Declare global functions that are available in the window scope
// eslint-disable-next-line no-var
declare var handleErr: (e: unknown) => void;
// eslint-disable-next-line no-var
declare var prettyErr: (data: { Fields?: Record<string, string> }) => void;

// Browser event types
interface BrowserEventTarget extends EventTarget {
  dataset: DOMStringMap;
  attributes: NamedNodeMap;
  nodeName: string;
  parentNode: Node | null;
  value: string;
  id: string;
}

interface BrowserEvent extends Event {
  target: BrowserEventTarget;
  currentTarget: BrowserEventTarget;
  preventDefault(): void;
}

interface InputEvent extends Event {
  target: BrowserEventTarget & HTMLInputElement;
}

// Window extension for debug purposes
interface Window {
  rootdev?: {
    invoiced?: unknown;
  };
}

// Common props interface for components that receive entity/year from routing
interface RouteProps {
  entity: string;
  year: string;
  id?: string;
  bucket?: string;
  key?: string;
}

// Props interface for components with parent reference
interface ParentProps<T> {
  parent: T;
}

// Export for use in modules
export {
  BrowserEventTarget,
  BrowserEvent,
  InputEvent,
  RouteProps,
  ParentProps,
};
