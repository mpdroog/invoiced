import * as React from "react";

export default class Dashboard extends React.Component<{}, {}> {
	render() {
		return <div className="normalheader">
		    <div className="hpanel">
		        <div className="panel-body">
		            <h2 className="font-light m-b-xs">
		                InvoiceD
		            </h2>
		            <small>General billing stuff here?</small>
		        </div>
		    </div>
		</div>;
	}
}
