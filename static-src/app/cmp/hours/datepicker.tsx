import * as React from "react";

export default class Datepicker extends React.Component<{}, {}> {
  constructor(props: any) {
    super(props);
    this.state = {
    	text: ""
    };
  }

  private update(e: InputEvent) {
    //console.log(e.target.value);
    let elem = e.target as any;

    if (elem.id === "import") {
      this.setState({text: e.target.value});
    }
  }

	private save(e: InputEvent) {
		e.preventDefault();
		this.props.importFn(
			importText(this.state.text.split("\n"))
		);
	}

	render() {
		console.log(this.props.value);
	    let selected = {fontSize:"18px"};
	    return (<div>
	    	<input type="date" value={this.props.value.format('YYYY-MM-DD')}/>
	        <div className="month"> 
	          <ul>
	            <li className="prev">&#10094;</li>
	            <li className="next">&#10095;</li>
	            <li>
	              August<br/>
	              <span style={selected}>{this.props.value.year()}</span>
	            </li>
	          </ul>
	        </div>

	        <ul className="weekdays">
	          <li>Mo</li>
	          <li>Tu</li>
	          <li>We</li>
	          <li>Th</li>
	          <li>Fr</li>
	          <li>Sa</li>
	          <li>Su</li>
	        </ul>

	        <ul className="days"> 
	          <li>1</li>
	          <li>2</li>
	          <li>3</li>
	          <li>4</li>
	          <li>5</li>
	          <li>6</li>
	          <li>7</li>
	          <li>8</li>
	          <li>9</li>
	          <li><span className="active">10</span></li>
	          <li>11</li>
	        </ul>
	    </div>);
	}
}