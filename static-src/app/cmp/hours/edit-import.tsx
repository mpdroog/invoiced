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

interface ImportResult {
	lines: ImportedLine[];
	errors: string[];
}

// Try parsing 2July, 1Aug into Date()
// @return Date|bool
function parseDate(line: string): string | false {
	line = line.replace(/ /g, "");
	const dayMonth = line.split(/^(\d{1,2})([a-zA-Z]+)$/g).filter((i: string) => i);
	if (dayMonth.length !== 2) {
		return false;
	}
	const day = dayMonth[0];
	const month = dayMonth[1];
	if (!day || !month) {
		return false;
	}
	const paddedDay = day.length === 1 ? "0" + day : day;
	const monthKey = month.toLowerCase() as MonthKey;

	const monthNum = monthMap[monthKey];
	if (!monthNum) {
		return false;
	}
	const dateStr = currentYear + "-" + monthNum + "-" + paddedDay;
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

function importText(lines: string[]): ImportResult {
	const output: ImportedLine[] = [];
	const errors: string[] = [];

	let day: string | null = null;
	let fromTo: string[][] = [];
	let text = "";
	let dayLineNum = 0;

	for (let i = 0; i < lines.length; i++) {
		const lineNum = i + 1;
		const rawLine = lines[i];
		if (!rawLine) continue;
		const line = rawLine.trim();
		if (line.length === 0) {
			// ignore empty lines
			continue;
		}

		const date = parseDate(line);
		if (date !== false) {
			if (day !== null) {
				// validate prev entry before saving
				if (fromTo.length === 0) {
					errors.push(`Line ${dayLineNum}: Date "${day}" has no time range`);
				}
				// save prev lines
				output.push({
					day: day,
					fromTo: fromTo,
					text: text.substring(3)
				});
			}

			// reset
			day = date;
			dayLineNum = lineNum;
			fromTo = [];
			text = "";
			continue;
		}

		const hours = parseHours(line);
		if (hours === false) {
			// text - but only if we have a date context
			if (day === null) {
				errors.push(`Line ${lineNum}: "${line}" appears before any date`);
			} else {
				text += " - " + line;
			}
		} else {
			// hours - but only if we have a date context
			if (day === null) {
				errors.push(`Line ${lineNum}: Time range "${line}" appears before any date`);
			} else {
				fromTo.push(hours);
			}
		}
	}

	// save last entry if we have one
	if (day !== null) {
		if (fromTo.length === 0) {
			errors.push(`Line ${dayLineNum}: Date "${day}" has no time range`);
		}
		output.push({
			day: day,
			fromTo: fromTo,
			text: text.substring(3)
		});
	}

	return { lines: output, errors };
}

interface IImportProps {
  importFn: (lines: ImportedLine[]) => void;
  hide: boolean;
  onHide: (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => void;
}

interface IImportState {
  text: string;
  errors: string[];
}

export default class HourImport extends React.Component<IImportProps, IImportState> {
  constructor(props: IImportProps) {
    super(props);
    this.state = {
    	text: "",
    	errors: []
    };
  }

  private update(e: React.ChangeEvent<HTMLTextAreaElement>): void {
    if (e.target.id === "import") {
      this.setState({text: e.target.value, errors: []});
    }
  }

	private save(e: React.MouseEvent<HTMLAnchorElement>): void {
		e.preventDefault();
		const result = importText(this.state.text.split("\n"));
		if (result.errors.length > 0) {
			this.setState({ errors: result.errors });
			return;
		}
		this.setState({ errors: [] });
		this.props.importFn(result.lines);
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
            <p className="text-muted"><small>Format: Start with a date (e.g. "27Mar"), then time ranges (e.g. "07:30 - 07:51"), then description text.</small></p>
		    <textarea id="import" style={t} onChange={this.update.bind(this)} placeholder={"27Mar\n07:30 - 12:00\nDescription of work done"}></textarea>
            {this.state.errors.length > 0 && (
              <div className="alert alert-danger" style={{marginTop: "10px"}}>
                <strong>Errors:</strong>
                <ul style={{marginBottom: 0, paddingLeft: "20px"}}>
                  {this.state.errors.map((err, i) => <li key={i}>{err}</li>)}
                </ul>
              </div>
            )}
          </div>
          <div className="modal-footer">
            <a onClick={this.save.bind(this)} className="btn btn-primary" style={{float:"right"}}> Parse</a>
          </div>
        </div>
      </div>
    </div>;
	}
}
