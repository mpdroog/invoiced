import * as React from "react";
import Axios from "axios";

export class InvoiceMail extends React.Component<{}, {}> {
  constructor(props) {
    super(props);
  }

  send(e) {
    let that = this;
    let parent = this.props.parent;
    let req = JSON.parse(JSON.stringify(parent.state.Mail));
    console.log("Send!", req);

    Axios.post('/api/v1/invoice/'+parent.props.entity+'/'+parent.props.year+'/'+parent.props.bucket+'/'+parent.state.Meta.Conceptid+'/email', req)
    .then(res => {
      parent.setState({Mail: that.state});
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
    let parent = this.props.parent;
    let parentState = this.props.parent.state;
    let hourFile = <div/>;
    if (parentState.Meta.HourFile.length > 0) {
      hourFile = <p><a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + parent.props.bucket + "/" + parentState.Meta.Conceptid + "/text"} target="_blank"><i className="fa fa-file-o" />&nbsp;hours.txt</a></p>;
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
	              <input type="text" className="form-control" value={parentState.Mail.Subject}/>
  	          	  <input type="text" className="form-control" value={"Reply-To: " + parentState.Mail.From} disabled={true}/>
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
		          	<input type="text" className="form-control" value={parentState.Mail.To}/>
		        </div>
		    </div>

            <textarea className="form-control h140">{parentState.Mail.Body}</textarea>
          </div>
          <div className="modal-footer">
            <div style={a}>
              <p><a href={"/api/v1/invoice/" + parent.props.entity + "/" + parent.props.year + "/" + parent.props.bucket + "/" + parentState.Meta.Conceptid + "/pdf"} target="_blank"><i className="fa fa-file-pdf-o" />&nbsp;{parentState.Meta.Invoiceid}.pdf</a></p>
              {hourFile}
            </div>
            <a onClick={this.send.bind(this)} className="btn btn-primary" style="float:right"> Send</a>
          </div>
        </div>
      </div>
    </div>;
  }
}