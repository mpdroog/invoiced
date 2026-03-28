import * as React from "react";

type MonthKey = 'jan' | 'feb' | 'mar' | 'apr' | 'may' | 'jun' | 'jul' | 'aug' | 'sep' | 'oct' | 'nov' | 'dec' |
	'january' | 'februari' | 'march' | 'april' | 'june' | 'juni' | 'july' | 'augustus' | 'september' | 'sept' | 'october' | 'okt' | 'november' | 'december';

const monthMap: Record<MonthKey, string> = {
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
	"sept": "09",
	"october": "10",
	"okt": "10",
	"november": "11",
	"december": "12",
}
const currentYear = new Date().getFullYear();

interface ImportedLine {
	day: string | null;
	fromTo: string[][];
	text: string;
}

// Try parsing 2July, 1Aug into Date()
// @return Date|bool
function parseDate(line: string): string | false {
	line = line.replace(/ /g, "");
	const dayMonth = line.split(/^(\d{1,2})([a-zA-Z]+)$/g).filter((i: string) => i);
	if (dayMonth.length !== 2) {
		return false;
	}
	if (dayMonth[0].length === 1) {
		dayMonth[0] = "0" + dayMonth[0];
	}
	const monthKey = dayMonth[1].toLowerCase() as MonthKey;

	const monthNum = monthMap[monthKey];
	if (!monthNum) {
		return false;
	}
	const dateStr = currentYear + "-" + monthNum + "-" + dayMonth[0];
	const date = new Date(dateStr);
	if (isNaN(date.getTime())) {
		return false;
	}

	return dateStr;
}
// Try parsing 12:08 - 13:12 into ["12:08", "13:12"]
// @return array|bool
function parseHours(line: string): string[] | false {
	line = line.replace(/ /g, "");
	const fromTo = line.split(/^([0-9]{2}:[0-9]{2})-([0-9]{2}:[0-9]{2})$/g).filter((i: string) => i);
	if (fromTo.length !== 2) {
		return false;
	}
	return fromTo;
}

function importText(lines: string[]): ImportedLine[] {
	const output: ImportedLine[] = [];

	let day: string | null = null;
	let fromTo: string[][] = [];
	let text = "";

	for (let i = 0; i < lines.length; i++) {
		const line = lines[i].trim();
		if (line.length === 0) {
			// ignore empty lines
			continue;
		}

		const date = parseDate(line);
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

		const hours = parseHours(line);
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

interface IImportProps {
  importFn: (lines: ImportedLine[]) => void;
  hide: boolean;
  onHide: (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => void;
}

interface IImportState {
  text: string;
}

export default class HourImport extends React.Component<IImportProps, IImportState> {
  constructor(props: IImportProps) {
    super(props);
    this.state = {
    	text: ""
    };
  }

  private update(e: React.ChangeEvent<HTMLTextAreaElement>): void {
    if (e.target.id === "import") {
      this.setState({text: e.target.value});
    }
  }

	private save(e: React.MouseEvent<HTMLAnchorElement>): void {
		e.preventDefault();
		this.props.importFn(
			importText(this.state.text.split("\n"))
		);
		this.props.onHide(e);
	}

	render(): React.JSX.Element {
		const s: React.CSSProperties = {display: "block"};
		const t: React.CSSProperties = {width:"100%", height: "200px"};
		if (! this.props.hide) {
			return <div/>;
		}

  	return <div className="modal" style={s} tabIndex={-1} role="dialog">
      <div className="modal-dialog">
        <div className="modal-content">
          <div className="modal-header">
            <button onClick={this.props.onHide} className="close" type="button" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true"> &times;</span>
            </button>
            <h4 className="modal-title">
              <i className="fa fa-arrow-up"></i>
              &nbsp;Import
            </h4>
          </div>
          <div className="modal-body">
		    <textarea id="import" style={t} onChange={this.update.bind(this)}></textarea>
          </div>
          <div className="modal-footer">
            <a onClick={this.save.bind(this)} className="btn btn-primary" style={{float:"right"}}> Parse</a>
          </div>
        </div>
      </div>
    </div>;
	}
}
