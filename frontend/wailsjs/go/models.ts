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
	export class ExperimentConditionFixOperationData {
	    operationId: string;
	    // Go type: time
	    fixedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentConditionFixOperationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationId = source["operationId"];
	        this.fixedAt = this.convertValues(source["fixedAt"], null);
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
	export class ExperimentPreparationPromptResponse {
	    sequenceNo: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentPreparationPromptResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sequenceNo = source["sequenceNo"];
	        this.content = source["content"];
	    }
	}
	export class ExperimentPreparationRequiredFieldsResponse {
	    purpose: boolean;
	    environmentConditions: boolean;
	    initialInput: boolean;
	    prompts: boolean;
	    evaluationAxes: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentPreparationRequiredFieldsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.purpose = source["purpose"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = source["prompts"];
	        this.evaluationAxes = source["evaluationAxes"];
	    }
	}
	export class ExperimentPreparationSourceResponse {
	    state: string;
	    versionId: string;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentPreparationSourceResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.versionId = source["versionId"];
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
	export class ExperimentWorkspaceEvaluationData {
	    id: string;
	    state: string;
	    summary?: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentWorkspaceEvaluationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.state = source["state"];
	        this.summary = source["summary"];
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
	export class ExperimentWorkspaceFixedConditionsData {
	    fixedConditionId: string;
	    purpose: string;
	    hypothesis?: string;
	    environmentConditions: string;
	    initialInput: string;
	    prompts: ExperimentPreparationPromptResponse[];
	    evaluationAxes: string;
	    // Go type: time
	    fixedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentWorkspaceFixedConditionsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fixedConditionId = source["fixedConditionId"];
	        this.purpose = source["purpose"];
	        this.hypothesis = source["hypothesis"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = this.convertValues(source["prompts"], ExperimentPreparationPromptResponse);
	        this.evaluationAxes = source["evaluationAxes"];
	        this.fixedAt = this.convertValues(source["fixedAt"], null);
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
	export class ExperimentWorkspaceRunData {
	    id: string;
	    state: string;
	    summary?: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentWorkspaceRunData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.state = source["state"];
	        this.summary = source["summary"];
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
	export class FixExperimentConditionsData {
	    experimentId: string;
	    state: string;
	    fixedConditionId: string;
	    operationId: string;
	    // Go type: time
	    fixedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new FixExperimentConditionsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	        this.fixedConditionId = source["fixedConditionId"];
	        this.operationId = source["operationId"];
	        this.fixedAt = this.convertValues(source["fixedAt"], null);
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
	export class FixExperimentConditionsError {
	    code: string;
	    message: string;
	    fieldErrors?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new FixExperimentConditionsError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.fieldErrors = source["fieldErrors"];
	    }
	}
	export class FixExperimentConditionsRequest {
	    requestId: string;
	    experimentId: string;
	    purpose: string;
	    hypothesis?: string;
	    environmentConditions: string;
	    initialInput: string;
	    prompts: string[];
	    evaluationAxes: string;
	
	    static createFrom(source: any = {}) {
	        return new FixExperimentConditionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	        this.purpose = source["purpose"];
	        this.hypothesis = source["hypothesis"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = source["prompts"];
	        this.evaluationAxes = source["evaluationAxes"];
	    }
	}
	export class FixExperimentConditionsResponse {
	    data?: FixExperimentConditionsData;
	    error?: FixExperimentConditionsError;
	
	    static createFrom(source: any = {}) {
	        return new FixExperimentConditionsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], FixExperimentConditionsData);
	        this.error = this.convertValues(source["error"], FixExperimentConditionsError);
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
	export class GetExperimentPreparationData {
	    experimentId: string;
	    state: string;
	    purpose: string;
	    hypothesis?: string;
	    environmentConditions: string;
	    initialInput: string;
	    prompts: ExperimentPreparationPromptResponse[];
	    evaluationAxes: string;
	    source: ExperimentPreparationSourceResponse;
	    requiredFields: ExperimentPreparationRequiredFieldsResponse;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentPreparationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	        this.purpose = source["purpose"];
	        this.hypothesis = source["hypothesis"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = this.convertValues(source["prompts"], ExperimentPreparationPromptResponse);
	        this.evaluationAxes = source["evaluationAxes"];
	        this.source = this.convertValues(source["source"], ExperimentPreparationSourceResponse);
	        this.requiredFields = this.convertValues(source["requiredFields"], ExperimentPreparationRequiredFieldsResponse);
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
	export class GetExperimentPreparationResponse {
	    data?: GetExperimentPreparationData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentPreparationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetExperimentPreparationData);
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
	export class GetExperimentWorkspaceData {
	    experimentId: string;
	    state: string;
	    fixedConditions: ExperimentWorkspaceFixedConditionsData;
	    conditionFixOperation: ExperimentConditionFixOperationData;
	    runs: ExperimentWorkspaceRunData[];
	    evaluations: ExperimentWorkspaceEvaluationData[];
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentWorkspaceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	        this.fixedConditions = this.convertValues(source["fixedConditions"], ExperimentWorkspaceFixedConditionsData);
	        this.conditionFixOperation = this.convertValues(source["conditionFixOperation"], ExperimentConditionFixOperationData);
	        this.runs = this.convertValues(source["runs"], ExperimentWorkspaceRunData);
	        this.evaluations = this.convertValues(source["evaluations"], ExperimentWorkspaceEvaluationData);
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
	export class GetExperimentWorkspaceResponse {
	    data?: GetExperimentWorkspaceData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentWorkspaceResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetExperimentWorkspaceData);
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
	export class RunDetailReconciliationData {
	    state: string;
	    // Go type: time
	    lastObservedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailReconciliationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.lastObservedAt = this.convertValues(source["lastObservedAt"], null);
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
	export class RunDetailFailureData {
	    code: string;
	    // Go type: time
	    occurredAt: any;
	    partialSummary?: string;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailFailureData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.occurredAt = this.convertValues(source["occurredAt"], null);
	        this.partialSummary = source["partialSummary"];
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
	export class RunDetailArtifactData {
	    digest: string;
	    label?: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailArtifactData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.digest = source["digest"];
	        this.label = source["label"];
	        this.status = source["status"];
	    }
	}
	export class RunDetailArtifactsData {
	    status: string;
	    items: RunDetailArtifactData[];
	    reasonCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailArtifactsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.items = this.convertValues(source["items"], RunDetailArtifactData);
	        this.reasonCode = source["reasonCode"];
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
	export class RunDetailObservationData {
	    sequenceNo: number;
	    kind: string;
	    // Go type: time
	    occurredAt: any;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailObservationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sequenceNo = source["sequenceNo"];
	        this.kind = source["kind"];
	        this.occurredAt = this.convertValues(source["occurredAt"], null);
	        this.summary = source["summary"];
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
	export class RunDetailOperationData {
	    id: string;
	    state: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailOperationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.state = source["state"];
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
	export class RunDetailRunData {
	    id: string;
	    experimentId: string;
	    state: string;
	    summary?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RunDetailRunData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	        this.summary = source["summary"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class GetRunDetailData {
	    run: RunDetailRunData;
	    fixedPrompt: ExperimentPreparationPromptResponse;
	    operation: RunDetailOperationData;
	    observations: RunDetailObservationData[];
	    artifacts: RunDetailArtifactsData;
	    failure?: RunDetailFailureData;
	    reconciliation: RunDetailReconciliationData;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetRunDetailData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.run = this.convertValues(source["run"], RunDetailRunData);
	        this.fixedPrompt = this.convertValues(source["fixedPrompt"], ExperimentPreparationPromptResponse);
	        this.operation = this.convertValues(source["operation"], RunDetailOperationData);
	        this.observations = this.convertValues(source["observations"], RunDetailObservationData);
	        this.artifacts = this.convertValues(source["artifacts"], RunDetailArtifactsData);
	        this.failure = this.convertValues(source["failure"], RunDetailFailureData);
	        this.reconciliation = this.convertValues(source["reconciliation"], RunDetailReconciliationData);
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
	export class GetRunDetailResponse {
	    data?: GetRunDetailData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetRunDetailResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetRunDetailData);
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
	export class PreparationResponse {
	    preparationId: string;
	    state: string;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    lastObservedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PreparationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preparationId = source["preparationId"];
	        this.state = source["state"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.lastObservedAt = this.convertValues(source["lastObservedAt"], null);
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
	export class ListPreparationsData {
	    preparations: PreparationResponse[];
	
	    static createFrom(source: any = {}) {
	        return new ListPreparationsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preparations = this.convertValues(source["preparations"], PreparationResponse);
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
	export class ListPreparationsResponse {
	    data?: ListPreparationsData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new ListPreparationsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], ListPreparationsData);
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
	
	
	
	
	
	
	
	
	
	export class SaveExperimentPreparationDraftData {
	    experimentId: string;
	    state: string;
	    purpose: string;
	    hypothesis?: string;
	    environmentConditions: string;
	    initialInput: string;
	    prompts: ExperimentPreparationPromptResponse[];
	    evaluationAxes: string;
	    // Go type: time
	    savedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SaveExperimentPreparationDraftData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.state = source["state"];
	        this.purpose = source["purpose"];
	        this.hypothesis = source["hypothesis"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = this.convertValues(source["prompts"], ExperimentPreparationPromptResponse);
	        this.evaluationAxes = source["evaluationAxes"];
	        this.savedAt = this.convertValues(source["savedAt"], null);
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
	export class SaveExperimentPreparationDraftRequest {
	    requestId: string;
	    experimentId: string;
	    purpose: string;
	    hypothesis?: string;
	    environmentConditions: string;
	    initialInput: string;
	    prompts: string[];
	    evaluationAxes: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveExperimentPreparationDraftRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	        this.purpose = source["purpose"];
	        this.hypothesis = source["hypothesis"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.prompts = source["prompts"];
	        this.evaluationAxes = source["evaluationAxes"];
	    }
	}
	export class SaveExperimentPreparationDraftResponse {
	    data?: SaveExperimentPreparationDraftData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new SaveExperimentPreparationDraftResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], SaveExperimentPreparationDraftData);
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
	export class StartExperimentData {
	    experimentId: string;
	    operationId: string;
	    state: string;
	    runs: ExperimentWorkspaceRunData[];
	
	    static createFrom(source: any = {}) {
	        return new StartExperimentData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.operationId = source["operationId"];
	        this.state = source["state"];
	        this.runs = this.convertValues(source["runs"], ExperimentWorkspaceRunData);
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
	export class StartExperimentRequest {
	    requestId: string;
	    experimentId: string;
	
	    static createFrom(source: any = {}) {
	        return new StartExperimentRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	    }
	}
	export class StartExperimentResponse {
	    data?: StartExperimentData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StartExperimentResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StartExperimentData);
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
	export class StartRunEvaluationData {
	    runId: string;
	    evaluationId: string;
	    operationId: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRunEvaluationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.evaluationId = source["evaluationId"];
	        this.operationId = source["operationId"];
	        this.state = source["state"];
	    }
	}
	export class StartRunEvaluationRequest {
	    requestId: string;
	    runId: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRunEvaluationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.runId = source["runId"];
	    }
	}
	export class StartRunEvaluationResponse {
	    data?: StartRunEvaluationData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StartRunEvaluationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StartRunEvaluationData);
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
	export class StopExperimentBriefingData {
	    operationId: string;
	
	    static createFrom(source: any = {}) {
	        return new StopExperimentBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationId = source["operationId"];
	    }
	}
	export class StopExperimentBriefingResponse {
	    data?: StopExperimentBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StopExperimentBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StopExperimentBriefingData);
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

