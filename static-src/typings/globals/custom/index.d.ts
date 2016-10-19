// Base element so we get dataset available
interface BrowserEventTarget extends EventTarget {
	dataset: DOMStringMap;
}
interface BrowserEvent extends Event {
    target: BrowserEventTarget;
}

// Base element so we get value on input elements
interface InputEventTarget extends BrowserEventTarget {
	value: string;
}
interface InputEvent extends BrowserEvent {
    target: InputEventTarget;
}

declare function handleErr(err: Error): void;
