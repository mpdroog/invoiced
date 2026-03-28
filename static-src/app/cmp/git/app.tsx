import * as React from "react";
import Axios from "axios";

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
  pushing: boolean;
  pulling: boolean;
  reverting: string | null;
  error: string | null;
  actionResult: string | null;
  status: StatusResponse | null;
  history: HistoryResponse | null;
  historyLoading: boolean;
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
      pushing: false,
      pulling: false,
      reverting: null,
      error: null,
      actionResult: null,
      status: null,
      history: null,
      historyLoading: false
    };
  }

  componentDidMount(): void {
    this.loadStatus();
    this.loadHistory(0);
  }

  private loadStatus(): void {
    this.setState({loading: true, error: null});
    Axios.get('/api/v1/git/' + this.props.entity + '/status')
      .then(res => {
        this.setState({loading: false, status: res.data});
      })
      .catch(err => {
        this.setState({loading: false, error: 'Failed to load git status'});
        handleErr(err);
      });
  }

  private loadHistory(page: number): void {
    this.setState({historyLoading: true});
    Axios.get('/api/v1/git/' + this.props.entity + '/history', {params: {page}})
      .then(res => {
        this.setState({historyLoading: false, history: res.data});
      })
      .catch(err => {
        this.setState({historyLoading: false});
        console.error('Failed to load history', err);
      });
  }

  private doPush(): void {
    this.setState({pushing: true, actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/push')
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({pushing: false, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({pushing: false, error: 'Push failed'});
        handleErr(err);
      });
  }

  private doPull(): void {
    this.setState({pulling: true, actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/pull')
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({pulling: false, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({pulling: false, error: 'Pull failed'});
        handleErr(err);
      });
  }

  private doDiscardAll(): void {
    if (!confirm('Discard ALL local changes? This will reset to the last pushed state.')) {
      return;
    }

    this.setState({reverting: 'all', actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/discard')
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({reverting: null, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({reverting: null, error: 'Discard failed'});
        handleErr(err);
      });
  }

  private doResetTo(fullHash: string, shortHash: string): void {
    if (!confirm(`Reset to commit ${shortHash}? Commits after this will be discarded.`)) {
      return;
    }

    this.setState({reverting: fullHash, actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/reset/' + fullHash)
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({reverting: null, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({reverting: null, error: 'Reset failed'});
        handleErr(err);
      });
  }

  render(): React.JSX.Element {
    let content = null;

    if (this.state.loading) {
      content = <p><i className="fa fa-spinner fa-spin"></i> Loading git status...</p>;
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
                  <button
                    className="btn btn-xs btn-default"
                    onClick={() => this.doResetTo(c.fullHash, c.hash)}
                    disabled={this.state.reverting !== null}
                    title="Reset to this commit"
                  >
                    {this.state.reverting === c.fullHash ? (
                      <i className="fa fa-spinner fa-spin"></i>
                    ) : (
                      <i className="fa fa-undo"></i>
                    )}
                  </button>
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
            <button
              className="btn btn-default m-r-sm"
              onClick={this.doPull.bind(this)}
              disabled={this.state.pulling || this.state.pushing || this.state.reverting !== null}
            >
              {this.state.pulling ? (
                <span><i className="fa fa-spinner fa-spin"></i> Pulling...</span>
              ) : (
                <span><i className="fa fa-cloud-download"></i> Pull from Remote</span>
              )}
            </button>
            <button
              className="btn btn-primary m-r-sm"
              onClick={this.doPush.bind(this)}
              disabled={this.state.pushing || this.state.pulling || this.state.reverting !== null || status.ahead === 0}
            >
              {this.state.pushing ? (
                <span><i className="fa fa-spinner fa-spin"></i> Pushing...</span>
              ) : (
                <span><i className="fa fa-cloud-upload"></i> Push to Remote</span>
              )}
            </button>
            {status.ahead > 0 && (
              <button
                className="btn btn-danger"
                onClick={this.doDiscardAll.bind(this)}
                disabled={this.state.pushing || this.state.pulling || this.state.reverting !== null}
              >
                {this.state.reverting === 'all' ? (
                  <span><i className="fa fa-spinner fa-spin"></i> Discarding...</span>
                ) : (
                  <span><i className="fa fa-undo"></i> Discard All</span>
                )}
              </button>
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
          <i className="fa fa-history"></i> Previous Commits
        </div>
        <div className="panel-body">
          {this.state.historyLoading ? (
            <p><i className="fa fa-spinner fa-spin"></i> Loading...</p>
          ) : (
            <div>
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
                <button
                  className="btn btn-default m-r-sm"
                  onClick={() => this.loadHistory(hist.page - 1)}
                  disabled={hist.page === 0 || this.state.historyLoading}
                >
                  <i className="fa fa-chevron-left"></i> Newer
                </button>
                <span className="text-muted">Page {hist.page + 1}</span>
                <button
                  className="btn btn-default m-l-sm"
                  onClick={() => this.loadHistory(hist.page + 1)}
                  disabled={!hist.hasMore || this.state.historyLoading}
                >
                  Older <i className="fa fa-chevron-right"></i>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>;
    }

    return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <i className="fa fa-git"></i> Git Status
          <div className="panel-tools">
            <button className="btn btn-default btn-xs" onClick={this.loadStatus.bind(this)}>
              <i className="fa fa-refresh"></i> Refresh
            </button>
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
