import * as React from "react";
import Axios from "axios";

interface LockedInputProps {
  locked: boolean;
  'data-key'?: string;
  type: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  value: string;
  placeholder?: string;
  id?: string;
  required?: boolean;
}

interface LockedInputState {
  locked: boolean;
}

/**
 * Lockable inputfield
 */
export class LockedInput extends React.Component<LockedInputProps, LockedInputState> {
  constructor(props: LockedInputProps) {
    super(props);
    this.state = {
      locked: props.locked
    }
  }

  private toggle(e: React.MouseEvent<HTMLAnchorElement>): void {
    e.preventDefault();
  	console.log("toggle lock");
  	this.setState({locked: !this.state.locked});
  }

  render(): React.JSX.Element {
    const locked = this.state.locked;
	const icon = locked ? "fa-lock" : "fa-unlock";
	return <div className="input-group">
		<input className="form-control" data-key={this.props['data-key']} disabled={locked} type={this.props.type} onChange={this.props.onChange} value={this.props.value} placeholder={this.props.placeholder} id={this.props.id} />
        <div className="input-group-addon">
        	<a onClick={this.toggle.bind(this)}>
        		<i className={"fa faa-ring animated-hover " + icon}></i>
        	</a>
        </div>
    </div>;
   }
}

interface AutocompleteProps {
  url: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onSelect: (data: AutocompleteSuggestion) => void;
  placeholder?: string;
  disabled?: boolean;
  required?: boolean;
  'data-key'?: string;
  id?: string;
}

interface AutocompleteSuggestion {
  Name: string;
  Street1?: string;
  Street2?: string;
  VAT?: string;
  COC?: string;
  NoteAdd?: string;
  BillingEmail?: string[];
  HourRate?: number;
}

interface AutocompleteState {
  suggestions: AutocompleteSuggestion[];
  show: boolean;
}

/**
 * Autocomplete dropdown
 */
export class Autocomplete extends React.Component<AutocompleteProps, AutocompleteState> {
  private lookupQuery: string = '';

  constructor(props: AutocompleteProps) {
    super(props);
    this.state = {
    	suggestions: [],
    	show: false
    }
  }

  lookup(e: React.ChangeEvent<HTMLInputElement>): void {
  	this.props.onChange(e);
    const txt = e.target.value;
    if (txt === this.lookupQuery) {
    	return;
    }
    this.lookupQuery = txt;

    console.log("lookup", this.props.url, txt);

    Axios.get(this.props.url, {params: {"query": txt}})
    .then(res => {
		console.log("lookup::suggest", res.data);
		this.setState({suggestions: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  onFocus(_e: React.FocusEvent<HTMLInputElement>): void {
    this.setState({show: true});
  }

  onBlur(e: React.FocusEvent<HTMLDivElement>): void {
    const currentTarget = e.currentTarget;
    window.setTimeout(() => {
      if (! currentTarget.contains(document.activeElement)) {
        this.setState({show: false});
      }
    }, 0);
  }

  onSelect(e: React.MouseEvent<HTMLAnchorElement>): void {
  	e.preventDefault();
  	const key = (e.target as HTMLAnchorElement).dataset["key"];
  	if (key === undefined) return;
  	const data = this.state.suggestions[parseInt(key, 10)];
  	if (!data) return;
	  this.props.onSelect(data);
    this.setState({show: false});
  }

  render(): React.JSX.Element {
	const p: React.CSSProperties = {
		position: "relative"
	};
	const s: React.CSSProperties = {
		position: "absolute",
		top: "30px",
		left: "0",
		right: "0",
		backgroundColor: "gray",
    zIndex: 2
	};
	const i: React.CSSProperties = {
		padding: "10px",
		backgroundColor: "#f9f9f9"
	}
	let suggest: React.JSX.Element | null = null;
	if (this.state.show && this.state.suggestions) {
		const items: React.JSX.Element[] = [];
		this.state.suggestions.forEach((item, idx) => {
			items.push(<div key={"suggest-"+idx} style={i}><a data-key={idx} onClick={this.onSelect.bind(this)}>{item.Name}</a></div>);
		})
		suggest = <div style={s}>
			{items}
		</div>;
	}

  let req: React.JSX.Element | null = null;
  if (this.props.required) {
    req = <i className="fa fa-asterisk text-danger fa-input"></i>;
  }
	return <div style={p} tabIndex={1} onBlur={this.onBlur.bind(this)}>
		<input type="text" className="form-control" onFocus={this.onFocus.bind(this)} onChange={this.lookup.bind(this)} value={this.props.value} disabled={this.props.disabled} placeholder={this.props.placeholder} data-key={this.props['data-key']} id={this.props.id} autoComplete="off" />
    {req}
		{suggest}
		</div>;
	}
}