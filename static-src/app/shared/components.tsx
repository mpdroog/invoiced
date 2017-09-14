import * as React from "react";
import Axios from "axios";

/**
 * Lockable inputfield
 */
export class LockedInput extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
    this.state = {
      locked: props.locked
    }
  }

  private toggle(e) {
    e.preventDefault();
  	console.log("toggle lock");
  	this.setState({locked: !this.state.locked});
  }

  render() {
    let locked = this.state.locked;
	let icon = locked ? "fa-lock" : "fa-unlock";
	return <div className="input-group">
		<input className="form-control" data-key={this.props['data-key']} disabled={locked} type={this.props.type} onChange={this.props.onChange} value={this.props.value} placeholder={this.props.placeholder}/>
        <div className="input-group-addon">
        	<a onClick={this.toggle.bind(this)}>
        		<i className={"fa faa-ring animated-hover " + icon}></i>
        	</a>
        </div>
    </div>;
   }
}

/**
 * Autocomplete dropdown
 */
export class Autocomplete extends React.Component<{}, {}> {
  private lookupQuery: string;

  constructor(props) {
    super(props);
    this.state = {
    	suggestions: [],
    	show: false
    }
  }

  lookup(e) {
  	this.props.onChange(e);
    let txt = e.target.value;
    if (txt === this.lookupQuery) {
    	return;
    }
    this.lookupQuery = txt;

    let that = this;
	console.log("lookup", this.props.url, txt);

    Axios.get(this.props.url, {params: {"query": txt}})
    .then(res => {
		console.log("lookup::suggest", res.data);
		that.setState({suggestions: res.data});
    })
    .catch(err => {
      handleErr(err);
    });
  }

  toggleDropdown() {
  	let that = this;
    window.setTimeout(function() {
	    that.setState({show: !that.state.show});
    }, 100);
  }

  onSelect(e) {
  	e.preventDefault();
  	let data = this.state.suggestions[ e.target.dataset["key"] ];
	this.props.onSelect(data);
  }

  render() {
	let p = {
		position: "relative"
	};
	let s = {
		position: "absolute",
		top: "30px",
		left: "0",
		right: "0",
		backgroundColor: "gray"
	};
	let i = {
		padding: "10px",
		backgroundColor: "#f9f9f9"
	}
	let suggest = null;
	if (this.state.show && this.state.suggestions) {
		let items = [];
		let that = this;
		this.state.suggestions.forEach(function(item, idx) {
			items.push(<div key={"suggest-"+idx} style={i}><a data-key={idx} onClick={that.onSelect.bind(that)}>{item.Name}</a></div>);
		})
		suggest = <div style={s}>
			{items}
		</div>;
	}

	return <div style={p}>
		<input type="text" className="form-control" onFocus={this.toggleDropdown.bind(this)} onBlur={this.toggleDropdown.bind(this)} onChange={this.lookup.bind(this)} value={this.props.value} disabled={this.props.disabled} data-key={this.props['data-key']} />
		{suggest}
		</div>;
	}
}