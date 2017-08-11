import * as React from "react";
import Invoices from "./list-bucket";

export default class InvoicesPage extends React.Component<{}, {}> {
	render() {
		return <div>
      <Invoices title="Pending Invoices" bucket="concepts" />
      <Invoices title="Pending Invoices" bucket="sales-invoices-unpaid" />
      <Invoices title="Paid Invoices" bucket="sales-invoices-paid" />
    </div>;
	}
}
