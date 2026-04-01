import * as React from "react";

// Simple shared backdrop - just listens for modal-open/modal-close events
export class ModalBackdrop extends React.Component<Record<string, never>, { count: number }> {
  constructor(props: Record<string, never>) {
    super(props);
    this.state = { count: 0 };
  }

  componentDidMount(): void {
    window.addEventListener("modal-open", this.onOpen);
    window.addEventListener("modal-close", this.onClose);
  }

  componentWillUnmount(): void {
    window.removeEventListener("modal-open", this.onOpen);
    window.removeEventListener("modal-close", this.onClose);
  }

  private onOpen = (): void => this.setState((s) => ({ count: s.count + 1 }));
  private onClose = (): void => this.setState((s) => ({ count: Math.max(0, s.count - 1) }));

  render(): React.JSX.Element | null {
    return this.state.count > 0 ? <div className="modal-backdrop in" /> : null;
  }
}

// Helper functions for any component to use
export function openModal(): void {
  window.dispatchEvent(new Event("modal-open"));
}

export function closeModal(): void {
  window.dispatchEvent(new Event("modal-close"));
}
