import * as React from "react";
import Axios from "axios";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import Moment from "moment";
import './dashboard.css';

// Chart colors - must match CSS variables in dashboard.css
const CHART_COLORS = {
	current: getComputedStyle(document.documentElement).getPropertyValue('--chart-color-current').trim() || '#62cb31',
	previous: getComputedStyle(document.documentElement).getPropertyValue('--chart-color-previous').trim() || '#3498db'
};

interface IDictionary {
	[index: string]: IMetricDay;
}

interface IMetricDay {
	RevenueTotal: string
	RevenueEx: string
	Hours: string
}

interface ICommit {
	hash: string
	fullHash: string
	message: string
	author: string
	date: string
}

interface IUnpaidSummary {
	Count: number
	TotalAmount: string
}

interface IOverdueInvoice {
	ID: string
	InvoiceID: string
	CustomerName: string
	DueDate: string
	Amount: string
	DaysOverdue: number
	Quarter: number
}

interface IQuarterSummary {
	Quarter: number
	InvoiceCount: number
	TotalRevenue: string
	TotalTax: string
	PaidCount: number
	UnpaidCount: number
	PaidRevenue: string
	UnpaidRevenue: string
}

interface IUnbilledHours {
	Count: number
	TotalHours: string
}

interface IYearComparison {
	CurrentYear: number
	PreviousYear: number
	CurrentRevenue: string
	PreviousRevenue: string
	GrowthPercent: string
	GrowthAmount: string
}

interface ITopClient {
	Name: string
	Revenue: string
	InvoiceCount: number
}

interface IDashboardData {
	monthly: IDictionary
	monthlyPrevYear: IDictionary
	unpaid: IUnpaidSummary
	overdue: IOverdueInvoice[]
	quarters: IQuarterSummary[]
	unbilledHours: IUnbilledHours
	yearComparison: IYearComparison
	topClients: ITopClient[]
}

interface IState {
	data: IDashboardData | null
	commits: ICommit[]
	loading: boolean
}

// Format number with space as thousands separator: 51868.65 -> 51 868,65
function formatCurrency(value: number): string {
	const parts = value.toFixed(2).split(".");
	const intPart = parts[0] ?? "0";
	parts[0] = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
	return parts.join(",");
}

interface DashboardProps {
	entity: string;
	year: string;
}

export default class Dashboard extends React.Component<DashboardProps, IState> {
	constructor(props: DashboardProps) {
		super(props);
		this.state = {
			data: null,
			commits: [],
			loading: true
		};
	}

	render(): React.JSX.Element {
		const { data, commits, loading } = this.state;
		const entity = this.props.entity;
		const year = this.props.year;

		if (loading || !data) {
			return <div className="normalheader text-center p-lg">
				<i className="fas fa-spinner fa-spin fa-2x"></i>
				<p>Loading dashboard...</p>
			</div>;
		}

		// Prepare chart data with year comparison
		const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
		const prevYear = parseInt(year) - 1;

		const revstats: {month: string, current: number, previous: number}[] = [];
		const hourstats: {month: string, current: number, previous: number}[] = [];
		let sum = 0;

		// Build data for each month (1-12)
		for (let m = 1; m <= 12; m++) {
			const monthStr = m.toString().padStart(2, '0');
			const currentKey = `${year}-${monthStr}`;
			const prevKey = `${prevYear}-${monthStr}`;

			// Current year
			const currentRev = data.monthly?.[currentKey] ? parseFloat(data.monthly[currentKey].RevenueEx) || 0 : 0;
			const currentHours = data.monthly?.[currentKey] ? parseFloat(data.monthly[currentKey].Hours) || 0 : 0;

			// Previous year
			const prevRev = data.monthlyPrevYear?.[prevKey] ? parseFloat(data.monthlyPrevYear[prevKey].RevenueEx) || 0 : 0;
			const prevHours = data.monthlyPrevYear?.[prevKey] ? parseFloat(data.monthlyPrevYear[prevKey].Hours) || 0 : 0;

			revstats.push({month: months[m-1] ?? "", current: currentRev, previous: prevRev});
			hourstats.push({month: months[m-1] ?? "", current: currentHours, previous: prevHours});

			sum += currentRev;
		}

		const growthPositive = data.yearComparison && parseFloat(data.yearComparison.GrowthAmount) >= 0;

		return <div className="normalheader">
			{/* Quick Stats Row */}
			<div className="row">
				<div className="col-md-3">
					<div className="hpanel stats">
						<div className="panel-body">
							<div className="stats-title">
								<span className="fas fa-money-bill"></span> Unpaid Invoices
							</div>
							<h1>&euro; {formatCurrency(parseFloat(data.unpaid?.TotalAmount || "0"))}</h1>
							<div className="stats-info">
								{data.unpaid?.Count || 0} invoices pending payment
							</div>
						</div>
					</div>
				</div>
				<div className="col-md-3">
					<div className="hpanel stats">
						<div className="panel-body">
							<div className="stats-title">
								<span className="far fa-clock"></span> Unbilled Hours
							</div>
							<h1>{formatCurrency(parseFloat(data.unbilledHours?.TotalHours || "0"))}</h1>
							<div className="stats-info">
								{data.unbilledHours?.Count || 0} hour sheets to bill
							</div>
						</div>
					</div>
				</div>
				<div className="col-md-3">
					<div className="hpanel stats">
						<div className="panel-body">
							<div className="stats-title">
								<span className="fas fa-chart-line"></span> Year Revenue
							</div>
							<h1>&euro; {formatCurrency(sum)}</h1>
							<div className="stats-info">
								{data.yearComparison && (
									<span className={growthPositive ? "text-success" : "text-danger"}>
										{growthPositive ? "+" : ""}{data.yearComparison.GrowthPercent}% vs {data.yearComparison.PreviousYear}
									</span>
								)}
							</div>
						</div>
					</div>
				</div>
				<div className="col-md-3">
					<div className="hpanel stats">
						<div className="panel-body">
							<div className="stats-title">
								<span className="fas fa-triangle-exclamation"></span> Overdue
							</div>
							<h1 className={data.overdue?.length > 0 ? "text-danger" : ""}>{data.overdue?.length || 0}</h1>
							<div className="stats-info">
								invoices past due date
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* Overdue Alerts */}
			{data.overdue && data.overdue.length > 0 && (
				<div className="row">
					<div className="col-md-12">
						<div className="hpanel hred">
							<div className="panel-heading">
								<i className="fas fa-triangle-exclamation"></i> Overdue Invoices
							</div>
							<div className="panel-body">
								<table className="table table-striped">
									<thead>
										<tr>
											<th>Invoice</th>
											<th>Customer</th>
											<th>Due Date</th>
											<th>Days Overdue</th>
											<th>Amount</th>
											<th></th>
										</tr>
									</thead>
									<tbody>
										{data.overdue.map((inv) => (
											<tr key={inv.ID}>
												<td><strong>{inv.InvoiceID}</strong></td>
												<td>{inv.CustomerName}</td>
												<td>{inv.DueDate}</td>
												<td>
													<span className={
														inv.DaysOverdue > 90 ? "label label-danger" :
														inv.DaysOverdue > 60 ? "label label-warning" :
														inv.DaysOverdue > 30 ? "label label-info" :
														"label label-default"
													}>
														{inv.DaysOverdue} days
													</span>
												</td>
												<td>&euro; {inv.Amount}</td>
												<td>
													<a href={`#${entity}/${year}/invoices/edit/Q${inv.Quarter}/${inv.ID}`} className="btn btn-xs btn-default">
														<i className="fas fa-eye"></i>
													</a>
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Main Content Row */}
			<div className="row">
				{/* Quarterly Breakdown */}
				<div className="col-md-6">
					<div className="hpanel">
						<div className="panel-heading">
							<i className="fas fa-calendar"></i> Quarterly Breakdown
						</div>
						<div className="panel-body">
							<table className="table">
								<thead>
									<tr>
										<th>Quarter</th>
										<th>Revenue</th>
										<th>Tax</th>
										<th>Paid / Unpaid</th>
									</tr>
								</thead>
								<tbody>
									{(data.quarters || []).map((q) => (
										<tr key={q.Quarter}>
											<td><strong>Q{q.Quarter}</strong></td>
											<td>&euro; {formatCurrency(parseFloat(q.TotalRevenue))}</td>
											<td>&euro; {formatCurrency(parseFloat(q.TotalTax))}</td>
											<td>
												<span className="text-success">{q.PaidCount}</span>
												{" / "}
												<span className="text-warning">{q.UnpaidCount}</span>
											</td>
										</tr>
									))}
									{(!data.quarters || data.quarters.length === 0) && (
										<tr><td colSpan={4} className="text-center text-muted">No data</td></tr>
									)}
								</tbody>
							</table>
						</div>
					</div>
				</div>

				{/* Top Clients */}
				<div className="col-md-6">
					<div className="hpanel">
						<div className="panel-heading">
							<i className="fas fa-users"></i> Top Clients
						</div>
						<div className="panel-body">
							<table className="table">
								<thead>
									<tr>
										<th>Client</th>
										<th>Revenue</th>
										<th>Invoices</th>
									</tr>
								</thead>
								<tbody>
									{(data.topClients || []).map((c, i) => (
										<tr key={i}>
											<td>{c.Name}</td>
											<td>&euro; {formatCurrency(parseFloat(c.Revenue))}</td>
											<td>{c.InvoiceCount}</td>
										</tr>
									))}
									{(!data.topClients || data.topClients.length === 0) && (
										<tr><td colSpan={3} className="text-center text-muted">No data</td></tr>
									)}
								</tbody>
							</table>
						</div>
					</div>
				</div>
			</div>

			{/* Charts Row */}
			<div className="row">
				<div className="col-md-6">
					<div className="hpanel">
						<div className="panel-heading">
							<i className="fas fa-chart-area"></i> Revenue Trend
							<span className="chart-legend pull-right">
								<span className="legend-current">{year}</span>
								<span className="legend-prev">{prevYear}</span>
							</span>
						</div>
						<div className="panel-body chart-container">
							<ResponsiveContainer width="100%" height="100%">
								<LineChart data={revstats}>
									<CartesianGrid strokeDasharray="3 3" />
									<XAxis dataKey="month" tick={{fontSize: 12}} interval={1} />
									<YAxis tick={{fontSize: 12}} />
									<Tooltip formatter={(value) => typeof value === 'number' ? `€ ${value.toFixed(2)}` : value} />
									<Line type="monotone" dataKey="current" stroke={CHART_COLORS.current} name={year} strokeWidth={2} dot={false} />
									<Line type="monotone" dataKey="previous" stroke={CHART_COLORS.previous} name={String(prevYear)} strokeWidth={2} dot={false} />
								</LineChart>
							</ResponsiveContainer>
						</div>
					</div>
				</div>

				<div className="col-md-6">
					<div className="hpanel">
						<div className="panel-heading">
							<i className="fas fa-chart-area"></i> Hours Trend
							<span className="chart-legend pull-right">
								<span className="legend-current">{year}</span>
								<span className="legend-prev">{prevYear}</span>
							</span>
						</div>
						<div className="panel-body chart-container">
							<ResponsiveContainer width="100%" height="100%">
								<LineChart data={hourstats}>
									<CartesianGrid strokeDasharray="3 3" />
									<XAxis dataKey="month" tick={{fontSize: 12}} interval={1} />
									<YAxis tick={{fontSize: 12}} />
									<Tooltip formatter={(value) => typeof value === 'number' ? `${value.toFixed(1)} hrs` : value} />
									<Line type="monotone" dataKey="current" stroke={CHART_COLORS.current} name={year} strokeWidth={2} dot={false} />
									<Line type="monotone" dataKey="previous" stroke={CHART_COLORS.previous} name={String(prevYear)} strokeWidth={2} dot={false} />
								</LineChart>
							</ResponsiveContainer>
						</div>
					</div>
				</div>
			</div>

			{/* Recent Activity */}
			<div className="row">
				<div className="col-md-12">
					<div className="hpanel">
						<div className="panel-heading">
							<i className="fas fa-history"></i> Recent Activity
						</div>
						<div className="panel-body">
							{commits.length === 0 ? (
								<p className="text-muted">No recent commits</p>
							) : (
								<div className="v-timeline vertical-container animate-panel" data-child="vertical-timeline-block" data-delay="1">
									{commits.slice(0, 5).map((commit) => (
										<div key={commit.hash} className="vertical-timeline-block">
											<div className="vertical-timeline-icon navy-bg">
												<i className="fas fa-calendar"></i>
											</div>
											<div className="vertical-timeline-content">
												<div className="p-sm">
													<span className="vertical-date pull-right">
														{Moment(commit.date).format('YYYY-MM-DD')} <br/>
														<small>{Moment(commit.date).format('HH:mm:ss')}</small>
													</span>
													<h2>{commit.message}</h2>
													<p>{Moment(commit.date).fromNow()}</p>
												</div>
												<div className="panel-footer">
													{commit.author}
												</div>
											</div>
										</div>
									))}
								</div>
							)}
						</div>
					</div>
				</div>
			</div>
		</div>;
	}

	componentDidMount(): void {
		this.ajax();
	}

	private ajax(): void {
		const entity = this.props.entity;
		const year = this.props.year;

		// Fetch comprehensive dashboard data
		Axios.get('/api/v1/dashboard/'+entity+'/'+year, {})
		.then(res => {
			this.setState({data: res.data, loading: false});
		})
		.catch(err => {
			console.error('Dashboard error:', err);
			this.setState({loading: false});
		});

		// Fetch git commits
		Axios.get('/api/v1/git/'+entity+'/history', {params: {page: 0}})
		.then(res => {
			this.setState({commits: res.data.commits || []});
		})
		.catch(_err => {
			// Silently ignore - git history may not be available
		});
	}
}
