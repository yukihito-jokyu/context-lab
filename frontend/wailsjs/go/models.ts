export namespace domain {
	
	export class ExperimentPreparationPrompt {
	    SequenceNo: number;
	    Content: string;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentPreparationPrompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SequenceNo = source["SequenceNo"];
	        this.Content = source["Content"];
	    }
	}
	export class DerivedExperimentChanges {
	    Purpose?: string;
	    Hypothesis?: string;
	    EnvironmentConditions?: string;
	    InitialInput?: string;
	    Prompts?: ExperimentPreparationPrompt[];
	    EvaluationAxes?: string;
	
	    static createFrom(source: any = {}) {
	        return new DerivedExperimentChanges(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Purpose = source["Purpose"];
	        this.Hypothesis = source["Hypothesis"];
	        this.EnvironmentConditions = source["EnvironmentConditions"];
	        this.InitialInput = source["InitialInput"];
	        this.Prompts = this.convertValues(source["Prompts"], ExperimentPreparationPrompt);
	        this.EvaluationAxes = source["EvaluationAxes"];
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

export namespace wails {
	
	export class CreateDerivedExperimentData {
	    requestId: string;
	    experimentId: string;
	    sourceExperimentId: string;
	    state: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDerivedExperimentData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	        this.sourceExperimentId = source["sourceExperimentId"];
	        this.state = source["state"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CreateDerivedExperimentRequest {
	    requestId: string;
	    sourceExperimentId: string;
	    changes: domain.DerivedExperimentChanges;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDerivedExperimentRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.sourceExperimentId = source["sourceExperimentId"];
	        this.changes = this.convertValues(source["changes"], domain.DerivedExperimentChanges);
	        this.reason = source["reason"];
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
	export class CreateDerivedExperimentResponse {
	    data?: CreateDerivedExperimentData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new CreateDerivedExperimentResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], CreateDerivedExperimentData);
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
	export class CreateInsightEvidenceData {
	    experimentId: string;
	    conclusionId: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateInsightEvidenceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.conclusionId = source["conclusionId"];
	    }
	}
	export class CreateInsightData {
	    requestId: string;
	    insightId: string;
	    evidences: CreateInsightEvidenceData[];
	    statement: string;
	    applicabilityConditions: string;
	    verificationGaps: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new CreateInsightData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.insightId = source["insightId"];
	        this.evidences = this.convertValues(source["evidences"], CreateInsightEvidenceData);
	        this.statement = source["statement"];
	        this.applicabilityConditions = source["applicabilityConditions"];
	        this.verificationGaps = source["verificationGaps"];
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
	
	export class CreateInsightEvidenceRequest {
	    experimentId: string;
	    conclusionId: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateInsightEvidenceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.conclusionId = source["conclusionId"];
	    }
	}
	export class CreateInsightRequest {
	    requestId: string;
	    evidences: CreateInsightEvidenceRequest[];
	    statement: string;
	    applicabilityConditions: string;
	    verificationGaps: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateInsightRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.evidences = this.convertValues(source["evidences"], CreateInsightEvidenceRequest);
	        this.statement = source["statement"];
	        this.applicabilityConditions = source["applicabilityConditions"];
	        this.verificationGaps = source["verificationGaps"];
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
	export class CreateInsightResponse {
	    data?: CreateInsightData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new CreateInsightResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], CreateInsightData);
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
	export class DerivationBriefingMessageResponse {
	    role: string;
	    content: string;
	    sequenceNo: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new DerivationBriefingMessageResponse(source);
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
	export class DerivationBriefingSuggestionResponse {
	    id: string;
	    versionNo: number;
	    purpose: string;
	    decision: string;
	    hypothesis?: string;
	    candidatePrompts: string[];
	    evaluationCriteria: string;
	    environmentConditions: string;
	    initialInput: string;
	    successCriteria: string;
	    requiredConditions: string;
	    openQuestion?: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new DerivationBriefingSuggestionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.versionNo = source["versionNo"];
	        this.purpose = source["purpose"];
	        this.decision = source["decision"];
	        this.hypothesis = source["hypothesis"];
	        this.candidatePrompts = source["candidatePrompts"];
	        this.evaluationCriteria = source["evaluationCriteria"];
	        this.environmentConditions = source["environmentConditions"];
	        this.initialInput = source["initialInput"];
	        this.successCriteria = source["successCriteria"];
	        this.requiredConditions = source["requiredConditions"];
	        this.openQuestion = source["openQuestion"];
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
	export class DerivationEligibilityData {
	    canCreateDerivedExperiment: boolean;
	    reasonCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new DerivationEligibilityData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canCreateDerivedExperiment = source["canCreateDerivedExperiment"];
	        this.reasonCode = source["reasonCode"];
	    }
	}
	export class ExperimentConclusionData {
	    id: string;
	    content: string;
	    state: string;
	    // Go type: time
	    finalizedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentConclusionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.state = source["state"];
	        this.finalizedAt = this.convertValues(source["finalizedAt"], null);
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
	export class DerivationSourceData {
	    experimentId: string;
	    purpose: string;
	    fixedConditions?: ExperimentWorkspaceFixedConditionsData;
	    conclusion?: ExperimentConclusionData;
	
	    static createFrom(source: any = {}) {
	        return new DerivationSourceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.purpose = source["purpose"];
	        this.fixedConditions = this.convertValues(source["fixedConditions"], ExperimentWorkspaceFixedConditionsData);
	        this.conclusion = this.convertValues(source["conclusion"], ExperimentConclusionData);
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
	
	export class EvaluationDetailEvaluationData {
	    id: string;
	    experimentId: string;
	    runId: string;
	    state: string;
	    summary?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailEvaluationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.experimentId = source["experimentId"];
	        this.runId = source["runId"];
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
	export class EvaluationDetailEvidenceData {
	    runSummary: string;
	    evaluationAxes: string;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailEvidenceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runSummary = source["runSummary"];
	        this.evaluationAxes = source["evaluationAxes"];
	    }
	}
	export class EvaluationDetailFailureData {
	    code: string;
	    // Go type: time
	    occurredAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailFailureData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.occurredAt = this.convertValues(source["occurredAt"], null);
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
	export class EvaluationDetailOperationData {
	    id: string;
	    state: string;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailOperationData(source);
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
	export class EvaluationDetailReconciliationData {
	    state: string;
	    // Go type: time
	    lastObservedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailReconciliationData(source);
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
	export class EvaluationDetailResultData {
	    status: string;
	    summary?: string;
	    reasonCode?: string;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationDetailResultData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.reasonCode = source["reasonCode"];
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
	export class ExperimentComparisonEvaluationData {
	    evaluationId: string;
	    runId: string;
	    state: string;
	    runSummary?: string;
	    result: EvaluationDetailResultData;
	    reconciliation: EvaluationDetailReconciliationData;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentComparisonEvaluationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.evaluationId = source["evaluationId"];
	        this.runId = source["runId"];
	        this.state = source["state"];
	        this.runSummary = source["runSummary"];
	        this.result = this.convertValues(source["result"], EvaluationDetailResultData);
	        this.reconciliation = this.convertValues(source["reconciliation"], EvaluationDetailReconciliationData);
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
	export class ExperimentComparisonExperimentData {
	    id: string;
	    purpose: string;
	    evaluationAxes: string;
	
	    static createFrom(source: any = {}) {
	        return new ExperimentComparisonExperimentData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.purpose = source["purpose"];
	        this.evaluationAxes = source["evaluationAxes"];
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
	
	export class ExperimentWorkspaceRunData {
	    id: string;
	    retryOfRunId?: string;
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
	        this.retryOfRunId = source["retryOfRunId"];
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
	export class FinalizeExperimentConclusionData {
	    requestId: string;
	    experimentId: string;
	    conclusionId: string;
	    conclusion: string;
	    state: string;
	    // Go type: time
	    finalizedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new FinalizeExperimentConclusionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	        this.conclusionId = source["conclusionId"];
	        this.conclusion = source["conclusion"];
	        this.state = source["state"];
	        this.finalizedAt = this.convertValues(source["finalizedAt"], null);
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
	export class FinalizeExperimentConclusionRequest {
	    requestId: string;
	    experimentId: string;
	    conclusion: string;
	
	    static createFrom(source: any = {}) {
	        return new FinalizeExperimentConclusionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.experimentId = source["experimentId"];
	        this.conclusion = source["conclusion"];
	    }
	}
	export class FinalizeExperimentConclusionResponse {
	    data?: FinalizeExperimentConclusionData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new FinalizeExperimentConclusionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], FinalizeExperimentConclusionData);
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
	export class GetDerivationBriefingData {
	    state: string;
	    messages: DerivationBriefingMessageResponse[];
	    latestSuggestion?: DerivationBriefingSuggestionResponse;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetDerivationBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.messages = this.convertValues(source["messages"], DerivationBriefingMessageResponse);
	        this.latestSuggestion = this.convertValues(source["latestSuggestion"], DerivationBriefingSuggestionResponse);
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
	export class GetDerivationBriefingResponse {
	    data?: GetDerivationBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetDerivationBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetDerivationBriefingData);
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
	export class GetDerivationSourceData {
	    source: DerivationSourceData;
	    eligibility: DerivationEligibilityData;
	
	    static createFrom(source: any = {}) {
	        return new GetDerivationSourceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = this.convertValues(source["source"], DerivationSourceData);
	        this.eligibility = this.convertValues(source["eligibility"], DerivationEligibilityData);
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
	export class GetDerivationSourceResponse {
	    data?: GetDerivationSourceData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetDerivationSourceResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetDerivationSourceData);
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
	export class GetEvaluationDetailData {
	    evaluation: EvaluationDetailEvaluationData;
	    operation: EvaluationDetailOperationData;
	    evidence: EvaluationDetailEvidenceData;
	    result: EvaluationDetailResultData;
	    failure?: EvaluationDetailFailureData;
	    reconciliation: EvaluationDetailReconciliationData;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetEvaluationDetailData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.evaluation = this.convertValues(source["evaluation"], EvaluationDetailEvaluationData);
	        this.operation = this.convertValues(source["operation"], EvaluationDetailOperationData);
	        this.evidence = this.convertValues(source["evidence"], EvaluationDetailEvidenceData);
	        this.result = this.convertValues(source["result"], EvaluationDetailResultData);
	        this.failure = this.convertValues(source["failure"], EvaluationDetailFailureData);
	        this.reconciliation = this.convertValues(source["reconciliation"], EvaluationDetailReconciliationData);
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
	export class GetEvaluationDetailResponse {
	    data?: GetEvaluationDetailData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetEvaluationDetailResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetEvaluationDetailData);
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
	export class GetExperimentComparisonData {
	    experiment: ExperimentComparisonExperimentData;
	    evaluations: ExperimentComparisonEvaluationData[];
	    conclusion?: ExperimentConclusionData;
	    // Go type: time
	    lastConfirmedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentComparisonData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experiment = this.convertValues(source["experiment"], ExperimentComparisonExperimentData);
	        this.evaluations = this.convertValues(source["evaluations"], ExperimentComparisonEvaluationData);
	        this.conclusion = this.convertValues(source["conclusion"], ExperimentConclusionData);
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
	export class GetExperimentComparisonResponse {
	    data?: GetExperimentComparisonData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetExperimentComparisonResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetExperimentComparisonData);
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
	export class InsightSummaryData {
	    id: string;
	    statement: string;
	    applicabilityConditions: string;
	    verificationGaps: string;
	    evidenceCount: number;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new InsightSummaryData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.statement = source["statement"];
	        this.applicabilityConditions = source["applicabilityConditions"];
	        this.verificationGaps = source["verificationGaps"];
	        this.evidenceCount = source["evidenceCount"];
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
	export class InsightSavedConsiderationData {
	    experimentId: string;
	    conclusionId: string;
	    content: string;
	    // Go type: time
	    finalizedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new InsightSavedConsiderationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.conclusionId = source["conclusionId"];
	        this.content = source["content"];
	        this.finalizedAt = this.convertValues(source["finalizedAt"], null);
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
	export class InsightEvidenceCandidateData {
	    experimentId: string;
	    purpose: string;
	    evaluationAxes: string;
	    conclusionId: string;
	    conclusion: string;
	    // Go type: time
	    finalizedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new InsightEvidenceCandidateData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.experimentId = source["experimentId"];
	        this.purpose = source["purpose"];
	        this.evaluationAxes = source["evaluationAxes"];
	        this.conclusionId = source["conclusionId"];
	        this.conclusion = source["conclusion"];
	        this.finalizedAt = this.convertValues(source["finalizedAt"], null);
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
	export class GetInsightWorkspaceData {
	    evidenceCandidates: InsightEvidenceCandidateData[];
	    savedConsiderations: InsightSavedConsiderationData[];
	    insights: InsightSummaryData[];
	    // Go type: time
	    lastConfirmedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new GetInsightWorkspaceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.evidenceCandidates = this.convertValues(source["evidenceCandidates"], InsightEvidenceCandidateData);
	        this.savedConsiderations = this.convertValues(source["savedConsiderations"], InsightSavedConsiderationData);
	        this.insights = this.convertValues(source["insights"], InsightSummaryData);
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
	export class GetInsightWorkspaceResponse {
	    data?: GetInsightWorkspaceData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetInsightWorkspaceResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetInsightWorkspaceData);
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
	export class PreparationReconciliationResponse {
	    state: string;
	    // Go type: time
	    lastObservedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PreparationReconciliationResponse(source);
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
	export class PreparationFailureResponse {
	    code: string;
	    // Go type: time
	    occurredAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PreparationFailureResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.occurredAt = this.convertValues(source["occurredAt"], null);
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
	export class PreparationDiagnosticResponse {
	    id: string;
	    code: string;
	    summary: string;
	    // Go type: time
	    occurredAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PreparationDiagnosticResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.summary = source["summary"];
	        this.occurredAt = this.convertValues(source["occurredAt"], null);
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
	export class PreparationCandidateResponse {
	    id: string;
	    environmentConditions: string;
	    summary: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new PreparationCandidateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.environmentConditions = source["environmentConditions"];
	        this.summary = source["summary"];
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
	export class GetPreparationData {
	    preparationId: string;
	    state: string;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    lastObservedAt: any;
	    candidates: PreparationCandidateResponse[];
	    diagnostics: PreparationDiagnosticResponse[];
	    failure?: PreparationFailureResponse;
	    reconciliation: PreparationReconciliationResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetPreparationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preparationId = source["preparationId"];
	        this.state = source["state"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.lastObservedAt = this.convertValues(source["lastObservedAt"], null);
	        this.candidates = this.convertValues(source["candidates"], PreparationCandidateResponse);
	        this.diagnostics = this.convertValues(source["diagnostics"], PreparationDiagnosticResponse);
	        this.failure = this.convertValues(source["failure"], PreparationFailureResponse);
	        this.reconciliation = this.convertValues(source["reconciliation"], PreparationReconciliationResponse);
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
	export class GetPreparationResponse {
	    data?: GetPreparationData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new GetPreparationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], GetPreparationData);
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
	
	
	
	
	
	
	export class RetryEndedRunData {
	    sourceRunId: string;
	    experimentId: string;
	    retryRunId: string;
	    operationId: string;
	    state: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RetryEndedRunData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceRunId = source["sourceRunId"];
	        this.experimentId = source["experimentId"];
	        this.retryRunId = source["retryRunId"];
	        this.operationId = source["operationId"];
	        this.state = source["state"];
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
	export class RetryEndedRunRequest {
	    requestId: string;
	    runId: string;
	
	    static createFrom(source: any = {}) {
	        return new RetryEndedRunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.runId = source["runId"];
	    }
	}
	export class RetryEndedRunResponse {
	    data?: RetryEndedRunData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new RetryEndedRunResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], RetryEndedRunData);
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
	export class SendDerivationBriefMessageData {
	    operationId: string;
	
	    static createFrom(source: any = {}) {
	        return new SendDerivationBriefMessageData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationId = source["operationId"];
	    }
	}
	export class SendDerivationBriefMessageResponse {
	    data?: SendDerivationBriefMessageData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new SendDerivationBriefMessageResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], SendDerivationBriefMessageData);
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
	export class StartDerivationBriefingData {
	    briefingSessionId: string;
	    operationId: string;
	    sourceExperimentId: string;
	
	    static createFrom(source: any = {}) {
	        return new StartDerivationBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.briefingSessionId = source["briefingSessionId"];
	        this.operationId = source["operationId"];
	        this.sourceExperimentId = source["sourceExperimentId"];
	    }
	}
	export class StartDerivationBriefingResponse {
	    data?: StartDerivationBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StartDerivationBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StartDerivationBriefingData);
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
	export class StartPreparationData {
	    preparationId: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new StartPreparationData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preparationId = source["preparationId"];
	        this.state = source["state"];
	    }
	}
	export class StartPreparationResponse {
	    data?: StartPreparationData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StartPreparationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StartPreparationData);
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
	export class StopDerivationBriefingData {
	    operationId: string;
	
	    static createFrom(source: any = {}) {
	        return new StopDerivationBriefingData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.operationId = source["operationId"];
	    }
	}
	export class StopDerivationBriefingResponse {
	    data?: StopDerivationBriefingData;
	    error?: ErrorResponse;
	
	    static createFrom(source: any = {}) {
	        return new StopDerivationBriefingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = this.convertValues(source["data"], StopDerivationBriefingData);
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

