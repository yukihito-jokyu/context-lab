export namespace wails {
	
	export class CreateExperimentFromBriefData {
	    experimentId: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateExperimentFromBriefData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	    }
	}
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
	export class CreateExperimentFromBriefResponse {
	    data?: CreateExperimentFromBriefData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new CreateExperimentFromBriefResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], CreateExperimentFromBriefData);
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
	
	export class ExperimentBriefResponse {
	    versionId: string;
	    purpose: string;
	    decision: string;
	    hypothesis?: string;
	    candidatePrompts: string[];
	    evaluationAxes: string;
	    environmentConditions: string;
	    initialInput: string;
	    successCriteria: string;
	    requiredConditions: string;
	    openQuestion?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentBriefResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.versionId = source["versionId"];
	        this.purpose = source["purpose"];
	        this.decision = source["decision"];
	        this.hypothesis = source["hypothesis"];
	        this.candidatePrompts = source["candidatePrompts"];
	        this.evaluationAxes = source["evaluationAxes"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.successCriteria = source["successCriteria"];
	        this.requiredConditions = source["requiredConditions"];
	        this.openQuestion = source["openQuestion"];
	    }
	}
	export class ExperimentMessageResponse {
	    role: string;
	    content: string;
	    sequenceNo: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentMessageResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.sequenceNo = source["sequenceNo"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class GetExperimentBriefingData {
	    state: string;
	    messages: ExperimentMessageResponse[];
	    latestBrief?: ExperimentBriefResponse;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.messages = this.convertValues(source["messages"], ExperimentMessageResponse);
	        this.latestBrief = this.convertValues(source["latestBrief"], ExperimentBriefResponse);
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
	export class GetExperimentBriefingResponse {
	    data?: GetExperimentBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetExperimentBriefingData);
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
	
	export class SendExperimentBriefMessageData {
	    operationId: string;
	
	    static createFrom(source: any = {}) {
	        return new SendExperimentBriefMessageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationId = source["operationId"];
	    }
	}
	export class SendExperimentBriefMessageResponse {
	    data?: SendExperimentBriefMessageData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new SendExperimentBriefMessageResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], SendExperimentBriefMessageData);
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

