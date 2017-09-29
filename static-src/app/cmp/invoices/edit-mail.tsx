import * as React from "react";
import Axios from "axios";

export class InvoiceMail extends React.Component<{}, {}> {
  constructor(props) {
    super(props);

    this.state = {
      From: "support",
      Body: `Dear customer,

Please find attached the latest invoice + hour specification.

With kind regards,
Mark Droog
RootDev`,
      Subject: "[SpyOFF] Invoice September-2017",
      To: "pw.droog@quicknet.nl"
    };
  }

  send(e) {
    let req = JSON.parse(JSON.stringify(this.state));
    console.log("Send!", req);

    let that = this;
    let parent = this.props.parent;
    Axios.post('/api/v1/invoice/'+parent.props.entity+'/'+parent.props.year+'/'+parent.props.bucket+'/'+parent.state.Meta.Conceptid+'/email', req)
    .then(res => {
      that.props.onHide(e);
    })
    .catch(err => {
      handleErr(err);
    });
  }

  render() {
    if (! this.props.hide) {
      return <div/>
    }
    let s = {display: "block"};
    let a = {float: "left", width: "350px", textAlign: "left"};
    let parentState = this.props.parent.state;
    let hourFile = <div/>;
    if (parentState.Meta.HourFile.length > 0) {
      hourFile = <p><a href={"/api/v1/invoice/rootdev/2017/Q3/" + parentState.Meta.Conceptid + "/text"} target="_blank"><i className="fa fa-file-o" />&nbsp;hours.txt</a></p>;
    }

  	return <div className="modal" style={s} tabindex="-1" role="dialog">
      <div className="modal-dialog">
        <div className="modal-content">
          <div className="modal-header">
            <button onClick={this.props.onHide} className="close" type="button" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true"> &times;</span>
            </button>
            <h4 className="modal-title">
            	<div className="row">
          		<div className="col-sm-1">
	              <i className="fa fa-send"></i>
          		</div>
	          	<div className="col-sm-10">
	              <input type="text" className="form-control" value={this.state.Subject}/>
  	          	  <input type="text" className="form-control" value="Reply-To: rootdev@gmail.com" disabled={true}/>
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
		          	<input type="text" className="form-control" value={this.state.To}/>
		        </div>
		    </div>

            <textarea className="form-control h140">{this.state.Body}</textarea>
          </div>
          <div className="modal-footer">
            <div style={a}>
              <p><a href={"/api/v1/invoice/rootdev/2017/Q3/" + parentState.Meta.Conceptid + "/pdf"} target="_blank"><i className="fa fa-file-pdf-o" />&nbsp;{parentState.Meta.Invoiceid}.pdf</a></p>
              {hourFile}
            </div>
            <a onClick={this.send.bind(this)} className="btn btn-primary" style="float:right"> Send</a>
          </div>
        </div>
      </div>
    </div>;
  }
}