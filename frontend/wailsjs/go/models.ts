export namespace shell {

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

}
