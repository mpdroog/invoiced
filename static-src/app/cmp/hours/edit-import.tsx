import * as React from "react";

const monthMap = {
	"jan": "01",
	"feb": "02",
	"mar": "03",
	"apr": "04",
	"may": "05",
	"jun": "06",
	"jul": "07",
	"aug": "08",
	"sep": "09",
	"oct": "10",
	"nov": "11",
	"dec": "12",

	"january": "01",
	"februari": "02",
	"march": "03",
	"april": "04",
	"june": "06",
	"juni": "06",
	"july": "07",
	"augustus": "08",
	"september": "09",
	"october": "10",
	"november": "11",
	"december": "12",
}

// Try parsing 2July, 1Aug into Date()
// @return Date|bool
function parseDate(line) {
	line = line.replace(/ /g, "");
	let dayMonth = line.split(/^(\d{1,2})([a-zA-Z]+)$/g).filter(i => i);
	if (dayMonth.length !== 2) {
		//console.log(line, dayMonth);
		return false;
	}
	if (dayMonth[0].length === 1) {
		dayMonth[0] = "0" + dayMonth[0];
	}
	dayMonth[1] = dayMonth[1].toLowerCase();

	let dateStr = "2017-" + monthMap[dayMonth[1]] + "-" + dayMonth[0];
	let date = new Date(dateStr);
	if (isNaN(date)) {
		//console.log("2017-" + monthMap[dayMonth[1]] + "-" + dayMonth[0]);
		return false;
	}

	return dateStr;
}
// Try parsing 12:08 - 13:12 into ["12:08", "13:12"]
// @return array|bool
function parseHours(line) {
	line = line.replace(/ /g, "");
	let fromTo = line.split(/^([0-9]{2}:[0-9]{2})-([0-9]{2}:[0-9]{2})$/g).filter(i => i);
	if (fromTo.length !== 2) {
		//console.log(fromTo.Length, fromTo);
		return false;
	}
	return fromTo;
}

function importText(lines) {
	let output = [];

	let day = null;
	let fromTo = [];
	let text = "";

	for (let i = 0; i < lines.length; i++) {
		let line = lines[i].trim();
		if (line.length === 0) {
			// ignore empty lines
			continue;
		}

		let date = parseDate(line);
		if (date !== false) {
			if (day !== null) {
				// save prev lines
				output.push({
					day: day,
					fromTo: fromTo,
					text: text.substr(3)
				});
			}

			// reset
			day = date;
			fromTo = [];
			text = "";
			continue;
		}

		let hours = parseHours(line);
		if (hours === false) {
			// text
			text += " - " + line;
		} else {
			// hours
			fromTo.push(hours);
		}
	}

	// save prev lines
	output.push({
		day: day,
		fromTo: fromTo,
		text: text.substr(3)
	});
	return output;
}

//importFn
interface IImportProps {
  importFn: func
}

export default class HourImport extends React.Component<IImportProps, {}> {
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
		var t = {width:"100%", height: "300px"};

		return <div className="normalheader"><div className="hpanel hblue">
			<div className="panel-heading hbuilt">
	          <div className="panel-tools">
	            <div className="btn-group nm7">
	              <button className="btn btn-default btn-hover-success" onClick={this.save.bind(this)}><i className="fa fa-arrow-up"></i>&nbsp;Import</button>
	            </div>
	          </div>
				Import hours
			</div>
			<div className="panel-body">
				<textarea id="import" style={t} onChange={this.update.bind(this)}></textarea>
			</div>
		</div></div>;
	}
}
