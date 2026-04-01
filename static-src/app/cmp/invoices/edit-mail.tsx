import * as React from "react";
import Axios from "axios";
import type {InvoiceMail as InvoiceMailType, InvoiceMeta} from "../../types/model";
import {ActionLink} from "../../shared/ActionButton";
import { openModal, closeModal } from "../../shared/Modal";

interface InvoiceMailParent {
  props: {
    entity: string;
    year: string;
  };
  state: {
    Mail?: InvoiceMailType;
    Meta?: InvoiceMeta;
    State: {
      currentBucket: string;
    };
  };
  setState: (state: {Mail: InvoiceMailType}) => void;
}

interface InvoiceMailProps {
  parent: InvoiceMailParent;
  hide: boolean;
  onHide: (e?: React.MouseEvent<HTMLButtonElement | HTMLAnchorElement>) => void;
}

interface InvoiceMailState {
  toError: string | null;
}

// Validate email format (simple check)
function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
}

// Validate comma-separated email list
function validateEmails(input: string): string | null {
  const trimmed = input.trim();
  if (trimmed === '') {
    return 'At least one email address is required';
  }
  // Split by comma, reject semicolons
  if (trimmed.includes(';')) {
    return 'Use commas to separate emails, not semicolons';
  }
  const parts = trimmed.split(',');
  for (const part of parts) {
    const email = part.trim();
    if (email === '') {
      return 'Empty email address (check for extra commas)';
    }
    if (!isValidEmail(email)) {
      return `Invalid email: ${email}`;
    }
  }
  return null;
}

export class InvoiceMail extends React.Component<InvoiceMailProps, InvoiceMailState> {
  constructor(props: InvoiceMailProps) {
    super(props);
    this.state = {
      toError: null
    };
  }

  componentDidMount(): void {
    if (this.props.hide) openModal();
    document.addEventListener('keydown', this.handleKeyDown);
  }

  componentDidUpdate(prevProps: InvoiceMailProps): void {
    if (!prevProps.hide && this.props.hide) openModal();
    if (prevProps.hide && !this.props.hide) closeModal();
  }

  componentWillUnmount(): void {
    if (this.props.hide) closeModal();
    document.removeEventListener('keydown', this.handleKeyDown);
  }

  private handleKeyDown = (e: KeyboardEvent): void => {
    if (e.key === 'Escape' && this.props.hide) {
      this.props.onHide();
    }
  };

  async send(): Promise<void> {
    const parent = this.props.parent;
    const mail = parent.state.Mail;
    const meta = parent.state.Meta;
    if (!mail || !meta) return;

    // Validate emails before sending
    const toError = validateEmails(mail.To);
    if (toError !== null) {
      this.setState({toError});
      return;
    }

    const req = structuredClone(mail);
    console.log("Send!", req);

    await Axios.post('/api/v1/invoice/'+parent.props.entity+'/'+parent.props.year+'/'+parent.state.State.currentBucket+'/'+meta.Conceptid+'/email', req);
    parent.setState({Mail: mail});
    // Close modal
    this.props.onHide();
  }

  private update(e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>): void {
    const id = e.target.id;
    const currentMail = this.props.parent.state.Mail;
    if (!currentMail) return;
    // Validate that id is a valid InvoiceMail key
    if (id !== 'From' && id !== 'Subject' && id !== 'To' && id !== 'Body') return;
    const mail = {...currentMail};
    mail[id] = e.target.value;
    this.props.parent.setState({Mail: mail});
    // Validate To field on every change
    if (id === 'To') {
      this.setState({toError: validateEmails(e.target.value)});
    }
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
    const bucket = parent.state.State.currentBucket;

    return <div className="modal modal-show" tabIndex={-1} role="dialog">
      <div className="modal-dialog modal-lg">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">
              <i className="fas fa-paper-plane me-2"></i>Send Invoice
            </h5>
            <button onClick={this.props.onHide} className="btn-close" type="button" aria-label="Close"></button>
          </div>
          <div className="modal-body">
            <div className="mb-3">
              <label htmlFor="To" className="form-label">To</label>
              <input
                type="text"
                className={`form-control ${this.state.toError !== null ? 'is-invalid' : ''}`}
                onChange={this.update.bind(this)}
                id="To"
                value={parentMail.To}
                placeholder="email@example.com, another@example.com"
              />
              {this.state.toError !== null && (
                <div className="invalid-feedback">{this.state.toError}</div>
              )}
              <div className="form-text">Separate multiple addresses with commas</div>
            </div>
            <div className="mb-3">
              <label htmlFor="Subject" className="form-label">Subject</label>
              <input type="text" className="form-control" onChange={this.update.bind(this)} id="Subject" value={parentMail.Subject}/>
            </div>
            <div className="mb-3">
              <label htmlFor="Body" className="form-label">Message</label>
              <textarea onChange={this.update.bind(this)} id="Body" className="form-control" rows={8}>{parentMail.Body}</textarea>
            </div>
            <div className="mb-0">
              <label className="form-label">Attachments</label>
              <div className="d-flex flex-wrap gap-3">
                <a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + bucket + "/" + parentMeta.Conceptid + "/pdf"} target="_blank" rel="noreferrer" className="btn btn-outline-secondary btn-sm">
                  <i className="far fa-file-pdf me-1"></i>{parentMeta.Invoiceid}.pdf
                </a>
                <a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + bucket + "/" + parentMeta.Conceptid + "/xml"} target="_blank" rel="noreferrer" className="btn btn-outline-secondary btn-sm">
                  <i className="fas fa-building-columns me-1"></i>{parentMeta.Invoiceid}.xml
                </a>
                {parentMeta.HourFile.length > 0 && (
                  <a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + bucket + "/" + parentMeta.Conceptid + "/text"} target="_blank" rel="noreferrer" className="btn btn-outline-secondary btn-sm">
                    <i className="far fa-file me-1"></i>hours.txt
                  </a>
                )}
              </div>
            </div>
          </div>
          <div className="modal-footer">
            <button onClick={this.props.onHide} className="btn btn-secondary" type="button">Cancel</button>
            <ActionLink onClick={this.send.bind(this)} className="btn btn-primary">
              <i className="fas fa-paper-plane me-1"></i>Send
            </ActionLink>
          </div>
        </div>
      </div>
    </div>;
  }
}