export namespace shell {

	export class AppError {
	    code: string;
	    severity: string;
	    retryable: boolean;
	    targetType?: string;
	    targetId?: string;
	    userMessage: string;
	    technicalDetail?: string;
	    recoveryActions: string[];
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new AppError(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.severity = source["severity"];
	        this.retryable = source["retryable"];
	        this.targetType = source["targetType"];
	        this.targetId = source["targetId"];
	        this.userMessage = source["userMessage"];
	        this.technicalDetail = source["technicalDetail"];
	        this.recoveryActions = source["recoveryActions"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class Health {
	    status: string;
	    severity: string;
	    userMessage: string;
	    technicalDetail: string;
	    recoveryActions: string[];
	    checkedAt: string;

	    static createFrom(source: any = {}) {
	        return new Health(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.severity = source["severity"];
	        this.userMessage = source["userMessage"];
	        this.technicalDetail = source["technicalDetail"];
	        this.recoveryActions = source["recoveryActions"];
	        this.checkedAt = source["checkedAt"];
	    }
	}
	export class Info {
	    appName: string;
	    stage: string;
	    shellVersion: string;
	    startedAt: string;
	    capabilities: string[];

	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.stage = source["stage"];
	        this.shellVersion = source["shellVersion"];
	        this.startedAt = source["startedAt"];
	        this.capabilities = source["capabilities"];
	    }
	}
	export class RuntimeEvent {
	    eventId: string;
	    runId?: string;
	    eventType: string;
	    state: string;
	    progress: number;
	    targetType?: string;
	    targetId?: string;
	    summary: string;
	    error?: AppError;
	    nextActions: string[];
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new RuntimeEvent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eventId = source["eventId"];
	        this.runId = source["runId"];
	        this.eventType = source["eventType"];
	        this.state = source["state"];
	        this.progress = source["progress"];
	        this.targetType = source["targetType"];
	        this.targetId = source["targetId"];
	        this.summary = source["summary"];
	        this.error = this.convertValues(source["error"], AppError);
	        this.nextActions = source["nextActions"];
	        this.createdAt = source["createdAt"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkbenchProbeCommand {
	    mode: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new WorkbenchProbeCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class WorkbenchStatus {
	    serviceName: string;
	    status: string;
	    summary: string;
	    capabilities: string[];
	    checkedAt: string;

	    static createFrom(source: any = {}) {
	        return new WorkbenchStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.capabilities = source["capabilities"];
	        this.checkedAt = source["checkedAt"];
	    }
	}
	export class WorkbenchProbeResult {
	    ok: boolean;
	    snapshot: WorkbenchStatus;
	    error?: AppError;
	    events: RuntimeEvent[];

	    static createFrom(source: any = {}) {
	        return new WorkbenchProbeResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.snapshot = this.convertValues(source["snapshot"], WorkbenchStatus);
	        this.error = this.convertValues(source["error"], AppError);
	        this.events = this.convertValues(source["events"], RuntimeEvent);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
