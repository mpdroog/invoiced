import * as React from "react";
import Invoices from "./invoices";

export default class InvoicesPage extends React.Component<{}, {}> {
	render() {
		return <div>
      <Invoices title="Pending Invoices" bucket="invoices" />
      <Invoices title="Paid Invoices" bucket="invoices-paid" />
    </div>;
	}
}
