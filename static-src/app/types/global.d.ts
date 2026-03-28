// Global type declarations for the invoiced application

// Declare global functions that are available in the window scope
declare function handleErr(e: unknown): void;
declare function prettyErr(data: { Fields?: Record<string, string> }): void;

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
