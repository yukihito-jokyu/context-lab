export namespace wails {
	
	export class ErrorResponse {
	    code: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ErrorResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class ExperimentResponse {
	    id: string;
	    purpose: string;
	    state: string;
	    progressSummary: string;
	    derivedFromExperimentId?: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.purpose = source["purpose"];
	        this.state = source["state"];
	        this.progressSummary = source["progressSummary"];
	        this.derivedFromExperimentId = source["derivedFromExperimentId"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class ResumeSummaryResponse {
	    recommendedExperimentId?: string;
	    statusCounts: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new ResumeSummaryResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recommendedExperimentId = source["recommendedExperimentId"];
	        this.statusCounts = source["statusCounts"];
	    }
	}
	export class ListExperimentsData {
	    experiments: ExperimentResponse[];
	    cancelledExperiments: ExperimentResponse[];
	    resumeSummary: ResumeSummaryResponse;
	    // Go type: time
	    lastConfirmedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ListExperimentsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experiments = this.convertValues(source["experiments"], ExperimentResponse);
	        this.cancelledExperiments = this.convertValues(source["cancelledExperiments"], ExperimentResponse);
	        this.resumeSummary = this.convertValues(source["resumeSummary"], ResumeSummaryResponse);
	        this.lastConfirmedAt = this.convertValues(source["lastConfirmedAt"], null);
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
	export class ListExperimentsResponse {
	    data?: ListExperimentsData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new ListExperimentsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], ListExperimentsData);
	        this.error = this.convertValues(source["error"], ErrorResponse);
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
	
	export class StartExperimentBriefingData {
	    briefingSessionId: string;
	    operationId: string;
	
	    static createFrom(source: any = {}) {
	        return new StartExperimentBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.briefingSessionId = source["briefingSessionId"];
	        this.operationId = source["operationId"];
	    }
	}
	export class StartExperimentBriefingResponse {
	    data?: StartExperimentBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StartExperimentBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StartExperimentBriefingData);
	        this.error = this.convertValues(source["error"], ErrorResponse);
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

