import * as React from "react";
import Axios from "axios";
import {ActionButton} from "../../shared/ActionButton";

interface CommitInfo {
  hash: string;
  fullHash: string;
  message: string;
  author: string;
  date: string;
}

interface StatusResponse {
  ahead: number;
  commits: CommitInfo[];
  remote: string;
}

interface HistoryResponse {
  commits: CommitInfo[];
  hasMore: boolean;
  page: number;
}

interface ActionResponse {
  success: boolean;
  message: string;
}

interface GitState {
  loading: boolean;
  error: string | null;
  actionResult: string | null;
  status: StatusResponse | null;
  history: HistoryResponse | null;
}

interface GitPageProps {
  entity: string;
  year: string;
}

export default class GitPage extends React.Component<GitPageProps, GitState> {
  constructor(props: GitPageProps) {
    super(props);
    this.state = {
      loading: true,
      error: null,
      actionResult: null,
      status: null,
      history: null
    };
  }

  componentDidMount(): void {
    void this.loadStatus();
    void this.loadHistory(0);
  }

  private async loadStatus(): Promise<void> {
    this.setState({loading: true, error: null});
    const res = await Axios.get('/api/v1/git/' + this.props.entity + '/status');
    this.setState({loading: false, status: res.data});
  }

  private async loadHistory(page: number): Promise<void> {
    const res = await Axios.get('/api/v1/git/' + this.props.entity + '/history', {params: {page}});
    this.setState({history: res.data});
  }

  private async doPush(): Promise<void> {
    this.setState({actionResult: null, error: null});
    const res = await Axios.post('/api/v1/git/' + this.props.entity + '/push');
    const result: ActionResponse = res.data;
    this.setState({actionResult: result.message});
    if (result.success) {
      await this.loadStatus();
    }
  }

  private async doPull(): Promise<void> {
    this.setState({actionResult: null, error: null});
    const res = await Axios.post('/api/v1/git/' + this.props.entity + '/pull');
    const result: ActionResponse = res.data;
    this.setState({actionResult: result.message});
    if (result.success) {
      await this.loadStatus();
    }
  }

  private async doDiscardAll(): Promise<void> {
    if (!confirm('Discard ALL local changes? This will reset to the last pushed state.')) {
      return;
    }

    this.setState({actionResult: null, error: null});
    const res = await Axios.post('/api/v1/git/' + this.props.entity + '/discard');
    const result: ActionResponse = res.data;
    this.setState({actionResult: result.message});
    if (result.success) {
      await this.loadStatus();
    }
  }

  private async doResetTo(fullHash: string, shortHash: string): Promise<void> {
    if (!confirm(`Reset to commit ${shortHash}? Commits after this will be discarded.`)) {
      return;
    }

    this.setState({actionResult: null, error: null});
    const res = await Axios.post('/api/v1/git/' + this.props.entity + '/reset/' + fullHash);
    const result: ActionResponse = res.data;
    this.setState({actionResult: result.message});
    if (result.success) {
      await this.loadStatus();
    }
  }

  render(): React.JSX.Element {
    let content = null;

    if (this.state.loading) {
      content = <p><i className="fas fa-spinner fa-spin"></i> Loading git status...</p>;
    } else if (this.state.error) {
      content = <div className="alert alert-danger">{this.state.error}</div>;
    } else if (this.state.status) {
      const status = this.state.status;

      let commitsList = null;
      if (status.commits.length > 0) {
        commitsList = <table className="table table-striped">
          <thead>
            <tr>
              <th>Hash</th>
              <th>Message</th>
              <th>Author</th>
              <th>Date</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {status.commits.map((c, i) => (
              <tr key={i}>
                <td><code>{c.hash}</code></td>
                <td>{c.message}</td>
                <td>{c.author}</td>
                <td>{c.date}</td>
                <td>
                  <ActionButton
                    className="btn btn-xs btn-default"
                    onClick={() => this.doResetTo(c.fullHash, c.hash)}
                    title="Reset to this commit"
                  >
                    <i className="fas fa-rotate-left"></i>
                  </ActionButton>
                </td>
              </tr>
            ))}
          </tbody>
        </table>;
      } else {
        commitsList = <p className="text-muted">No unpushed commits. Everything is up to date.</p>;
      }

      content = <div>
        <div className="row m-b-md">
          <div className="col-md-6">
            <p><strong>Remote:</strong> {status.remote || 'Not configured'}</p>
            <p><strong>Commits ahead:</strong> {status.ahead}</p>
          </div>
          <div className="col-md-6 text-right">
            <ActionButton
              className="btn btn-default m-r-sm"
              onClick={this.doPull.bind(this)}
            >
              <i className="fas fa-cloud-arrow-down"></i> Pull from Remote
            </ActionButton>
            <ActionButton
              className="btn btn-primary m-r-sm"
              onClick={this.doPush.bind(this)}
              disabled={status.ahead === 0}
            >
              <i className="fas fa-cloud-arrow-up"></i> Push to Remote
            </ActionButton>
            {status.ahead > 0 && (
              <ActionButton
                className="btn btn-danger"
                onClick={this.doDiscardAll.bind(this)}
              >
                <i className="fas fa-rotate-left"></i> Discard All
              </ActionButton>
            )}
          </div>
        </div>

        {this.state.actionResult && (
          <div className="alert alert-info">{this.state.actionResult}</div>
        )}

        <h4>Unpushed Commits</h4>
        {commitsList}
      </div>;
    }

    let historyPanel = null;
    if (this.state.history) {
      const hist = this.state.history;
      historyPanel = <div className="hpanel hgreen m-t-md">
        <div className="panel-heading hbuilt">
          <i className="fas fa-history"></i> Previous Commits
        </div>
        <div className="panel-body">
          <table className="table table-striped">
            <thead>
              <tr>
                <th>Hash</th>
                <th>Message</th>
                <th>Author</th>
                <th>Date</th>
              </tr>
            </thead>
            <tbody>
              {hist.commits.map((c, i) => (
                <tr key={i}>
                  <td><code>{c.hash}</code></td>
                  <td>{c.message}</td>
                  <td>{c.author}</td>
                  <td>{c.date}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="text-center">
            <ActionButton
              className="btn btn-default m-r-sm"
              onClick={() => this.loadHistory(hist.page - 1)}
              disabled={hist.page === 0}
            >
              <i className="fas fa-chevron-left"></i> Newer
            </ActionButton>
            <span className="text-muted">Page {hist.page + 1}</span>
            <ActionButton
              className="btn btn-default m-l-sm"
              onClick={() => this.loadHistory(hist.page + 1)}
              disabled={!hist.hasMore}
            >
              Older <i className="fas fa-chevron-right"></i>
            </ActionButton>
          </div>
        </div>
      </div>;
    }

    return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <i className="fab fa-git"></i> Git Status
          <div className="panel-tools">
            <ActionButton className="btn btn-default btn-xs" onClick={this.loadStatus.bind(this)}>
              <i className="fas fa-rotate"></i> Refresh
            </ActionButton>
          </div>
        </div>
        <div className="panel-body">
          {content}
        </div>
      </div>
      {historyPanel}
    </div>;
  }
}
