import * as React from "react";
import Axios from "axios";
import {IInvoiceMail, IInvoiceMeta} from "./edit-struct";
import {ActionLink} from "../../shared/ActionButton";
import { openModal, closeModal } from "../../shared/Modal";

interface InvoiceMailParent {
  props: {
    entity: string;
    year: string;
    bucket?: string;
  };
  state: {
    Mail?: IInvoiceMail;
    Meta?: IInvoiceMeta;
  };
  setState: (state: {Mail: IInvoiceMail}) => void;
}

interface InvoiceMailProps {
  parent: InvoiceMailParent;
  hide: boolean;
  onHide: (e: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => void;
}

export class InvoiceMail extends React.Component<InvoiceMailProps, Record<string, never>> {
  constructor(props: InvoiceMailProps) {
    super(props);
  }

  componentDidMount(): void {
    if (this.props.hide) openModal();
  }

  componentDidUpdate(prevProps: InvoiceMailProps): void {
    if (!prevProps.hide && this.props.hide) openModal();
    if (prevProps.hide && !this.props.hide) closeModal();
  }

  componentWillUnmount(): void {
    if (this.props.hide) closeModal();
  }

  async send(): Promise<void> {
    const parent = this.props.parent;
    const mail = parent.state.Mail;
    const meta = parent.state.Meta;
    if (!mail || !meta) return;
    const req = JSON.parse(JSON.stringify(mail));
    console.log("Send!", req);

    await Axios.post('/api/v1/invoice/'+parent.props.entity+'/'+parent.props.year+'/'+parent.props.bucket+'/'+meta.Conceptid+'/email', req);
    parent.setState({Mail: mail});
    // Close modal by simulating click event
    this.props.onHide({} as React.MouseEvent<HTMLAnchorElement>);
  }

  private update(e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>): void {
    const id = e.target.id as keyof IInvoiceMail;
    const currentMail = this.props.parent.state.Mail;
    if (!currentMail) return;
    const mail = {...currentMail};
    mail[id] = e.target.value;
    this.props.parent.setState({Mail: mail});
  }

  render(): React.JSX.Element {
    if (! this.props.hide) {
      return <div/>
    }
    const parent = this.props.parent;
    const parentMail = parent.state.Mail;
    const parentMeta = parent.state.Meta;
    if (!parentMail || !parentMeta) {
      return <div/>;
    }
    const s: React.CSSProperties = {display: "block"};
    let hourFile: React.JSX.Element = <div/>;
    if (parentMeta.HourFile.length > 0) {
      hourFile = <p><a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + parent.props.bucket + "/" + parentMeta.Conceptid + "/text"} target="_blank" rel="noreferrer"><i className="far fa-file" />&nbsp;hours.txt</a></p>;
    }

  	return <div className="modal" style={s} tabIndex={-1} role="dialog">
      <div className="modal-dialog">
        <div className="modal-content">
          <div className="modal-header">
            <button onClick={this.props.onHide} className="close" type="button" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true"> &times;</span>
            </button>
            <h4 className="modal-title">
            	<div className="row">
          		<div className="col-sm-1">
	              <i className="fas fa-paper-plane"></i>
          		</div>
	          	<div className="col-sm-10">
	              <input type="text" className="form-control" onChange={this.update.bind(this)} id="Subject" value={parentMail.Subject}/>
  	          	  <input type="text" className="form-control" onChange={this.update.bind(this)} id="From" value={"Reply-To: " + parentMail.From} disabled={true}/>
	            </div>
	            </div>
            </h4>
          </div>
          <div className="modal-body">
		    <div className="row">
          		<div className="col-sm-1 pt8">
          			To
          		</div>
	          	<div className="col-sm-11">
		          	<input type="text" className="form-control" onChange={this.update.bind(this)} id="To" value={parentMail.To}/>
		        </div>
		    </div>

            <textarea onChange={this.update.bind(this)} id="Body" className="form-control h140">{parentMail.Body}</textarea>
          </div>
          <div className="modal-footer">
            <div className="email-attachments">
              <p><a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + parent.props.bucket + "/" + parentMeta.Conceptid + "/pdf"} target="_blank" rel="noreferrer"><i className="far fa-file-pdf" />&nbsp;{parentMeta.Invoiceid}.pdf</a></p>
              <p><a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + parent.props.bucket + "/" + parentMeta.Conceptid + "/xml"} target="_blank" rel="noreferrer"><i className="fas fa-building-columns" />&nbsp;{parentMeta.Invoiceid}.xml</a></p>
              {hourFile}
            </div>
            <ActionLink onClick={this.send.bind(this)} className="btn btn-primary pull-right"> Send</ActionLink>
          </div>
        </div>
      </div>
    </div>;
  }
}