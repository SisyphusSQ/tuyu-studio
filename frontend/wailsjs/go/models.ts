export namespace project {

	export class CanvasGridDTO {
	    visible: boolean;
	    size: number;
	    opacity: number;

	    static createFrom(source: any = {}) {
	        return new CanvasGridDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.visible = source["visible"];
	        this.size = source["size"];
	        this.opacity = source["opacity"];
	    }
	}
	export class CanvasLayoutDTO {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	    collapsed: boolean;

	    static createFrom(source: any = {}) {
	        return new CanvasLayoutDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.collapsed = source["collapsed"];
	    }
	}
	export class CanvasPositionDTO {
	    x: number;
	    y: number;

	    static createFrom(source: any = {}) {
	        return new CanvasPositionDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class CanvasSizeDTO {
	    width: number;
	    height: number;

	    static createFrom(source: any = {}) {
	        return new CanvasSizeDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class CanvasViewportDTO {
	    x: number;
	    y: number;
	    zoom: number;

	    static createFrom(source: any = {}) {
	        return new CanvasViewportDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.zoom = source["zoom"];
	    }
	}
	export class CheckProjectHealthCommand {
	    root: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new CheckProjectHealthCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class ScriptSourceRange {
	    startLine: number;
	    endLine: number;

	    static createFrom(source: any = {}) {
	        return new ScriptSourceRange(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startLine = source["startLine"];
	        this.endLine = source["endLine"];
	    }
	}
	export class DialogueLineDTO {
	    characterName: string;
	    text: string;
	    intent?: string;

	    static createFrom(source: any = {}) {
	        return new DialogueLineDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterName = source["characterName"];
	        this.text = source["text"];
	        this.intent = source["intent"];
	    }
	}
	export class ConfirmScriptSceneCommand {
	    root: string;
	    scriptId?: string;
	    sceneId?: string;
	    title: string;
	    location: string;
	    timeOfDay?: string;
	    characters: string[];
	    props: string[];
	    action: string;
	    dialogue: DialogueLineDTO[];
	    emotionalBeat?: string;
	    sourceRange: ScriptSourceRange;
	    allowOverlap: boolean;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new ConfirmScriptSceneCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.sceneId = source["sceneId"];
	        this.title = source["title"];
	        this.location = source["location"];
	        this.timeOfDay = source["timeOfDay"];
	        this.characters = source["characters"];
	        this.props = source["props"];
	        this.action = source["action"];
	        this.dialogue = this.convertValues(source["dialogue"], DialogueLineDTO);
	        this.emotionalBeat = source["emotionalBeat"];
	        this.sourceRange = this.convertValues(source["sourceRange"], ScriptSourceRange);
	        this.allowOverlap = source["allowOverlap"];
	        this.correlationId = source["correlationId"];
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
	export class ConfirmShotCandidateCommand {
	    root: string;
	    scriptId?: string;
	    candidateId: string;
	    shotId?: string;
	    confirmedBy?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new ConfirmShotCandidateCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.candidateId = source["candidateId"];
	        this.shotId = source["shotId"];
	        this.confirmedBy = source["confirmedBy"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class CreateProjectCommand {
	    root: string;
	    projectId: string;
	    name: string;
	    type: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new CreateProjectCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.correlationId = source["correlationId"];
	    }
	}

	export class FrameHistorySummaryDTO {
	    currentRunId?: string;
	    favoriteRunIds: string[];
	    latestSuccessfulRunId?: string;

	    static createFrom(source: any = {}) {
	        return new FrameHistorySummaryDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentRunId = source["currentRunId"];
	        this.favoriteRunIds = source["favoriteRunIds"];
	        this.latestSuccessfulRunId = source["latestSuccessfulRunId"];
	    }
	}
	export class OperationError {
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
	        return new OperationError(source);
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
	export class GraphEdgeDTO {
	    id: string;
	    sourceNodeId: string;
	    targetNodeId: string;
	    relation: string;
	    label: string;
	    createdAt: string;
	    validity: string;
	    error?: OperationError;

	    static createFrom(source: any = {}) {
	        return new GraphEdgeDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceNodeId = source["sourceNodeId"];
	        this.targetNodeId = source["targetNodeId"];
	        this.relation = source["relation"];
	        this.label = source["label"];
	        this.createdAt = source["createdAt"];
	        this.validity = source["validity"];
	        this.error = this.convertValues(source["error"], OperationError);
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
	export class GraphNodeDTO {
	    id: string;
	    kind: string;
	    category: string;
	    title: string;
	    refId?: string;
	    position: CanvasPositionDTO;
	    size: CanvasSizeDTO;
	    collapsed: boolean;
	    status?: string;
	    badges: string[];
	    source: string;
	    sourceEventId?: string;
	    data?: Record<string, string>;

	    static createFrom(source: any = {}) {
	        return new GraphNodeDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.category = source["category"];
	        this.title = source["title"];
	        this.refId = source["refId"];
	        this.position = this.convertValues(source["position"], CanvasPositionDTO);
	        this.size = this.convertValues(source["size"], CanvasSizeDTO);
	        this.collapsed = source["collapsed"];
	        this.status = source["status"];
	        this.badges = source["badges"];
	        this.source = source["source"];
	        this.sourceEventId = source["sourceEventId"];
	        this.data = source["data"];
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
	export class GraphNodeLayoutCommand {
	    id: string;
	    position: CanvasPositionDTO;
	    size?: CanvasSizeDTO;
	    collapsed: boolean;

	    static createFrom(source: any = {}) {
	        return new GraphNodeLayoutCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.position = this.convertValues(source["position"], CanvasPositionDTO);
	        this.size = this.convertValues(source["size"], CanvasSizeDTO);
	        this.collapsed = source["collapsed"];
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
	export class GraphViewCommand {
	    root: string;
	    expectedGraphVersion?: number;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new GraphViewCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.expectedGraphVersion = source["expectedGraphVersion"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class ProjectEvent {
	    eventId: string;
	    eventType: string;
	    state: string;
	    summary: string;
	    error?: OperationError;
	    nextActions: string[];
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new ProjectEvent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eventId = source["eventId"];
	        this.eventType = source["eventType"];
	        this.state = source["state"];
	        this.summary = source["summary"];
	        this.error = this.convertValues(source["error"], OperationError);
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
	export class HealthItem {
	    severity: string;
	    code: string;
	    path?: string;
	    affectedObjects: string[];
	    userMessage: string;
	    technicalDetail?: string;
	    recoveryActions: string[];

	    static createFrom(source: any = {}) {
	        return new HealthItem(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.code = source["code"];
	        this.path = source["path"];
	        this.affectedObjects = source["affectedObjects"];
	        this.userMessage = source["userMessage"];
	        this.technicalDetail = source["technicalDetail"];
	        this.recoveryActions = source["recoveryActions"];
	    }
	}
	export class HealthReport {
	    status: string;
	    checkedAt: string;
	    items: HealthItem[];

	    static createFrom(source: any = {}) {
	        return new HealthReport(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.checkedAt = source["checkedAt"];
	        this.items = this.convertValues(source["items"], HealthItem);
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
	export class ReferenceGroupDTO {
	    id: string;
	    title: string;
	    role: string;
	    inputNodeIds: string[];
	    priority: number;
	    notes?: string;
	    layout: CanvasLayoutDTO;

	    static createFrom(source: any = {}) {
	        return new ReferenceGroupDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.role = source["role"];
	        this.inputNodeIds = source["inputNodeIds"];
	        this.priority = source["priority"];
	        this.notes = source["notes"];
	        this.layout = this.convertValues(source["layout"], CanvasLayoutDTO);
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
	export class ProductionFrameDTO {
	    id: string;
	    title: string;
	    referenceGroupIds: string[];
	    outputNodeIds: string[];
	    taskIntent: string;
	    requiredCapabilities: string[];
	    historySummary: FrameHistorySummaryDTO;
	    layout: CanvasLayoutDTO;

	    static createFrom(source: any = {}) {
	        return new ProductionFrameDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.referenceGroupIds = source["referenceGroupIds"];
	        this.outputNodeIds = source["outputNodeIds"];
	        this.taskIntent = source["taskIntent"];
	        this.requiredCapabilities = source["requiredCapabilities"];
	        this.historySummary = this.convertValues(source["historySummary"], FrameHistorySummaryDTO);
	        this.layout = this.convertValues(source["layout"], CanvasLayoutDTO);
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
	export class ProjectCanvasDTO {
	    id: string;
	    projectId: string;
	    schemaVersion: string;
	    version: number;
	    viewport: CanvasViewportDTO;
	    theme: string;
	    grid: CanvasGridDTO;
	    nodes: GraphNodeDTO[];
	    edges: GraphEdgeDTO[];
	    frames: ProductionFrameDTO[];
	    referenceGroups: ReferenceGroupDTO[];
	    selectedNodeIds: string[];
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new ProjectCanvasDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.schemaVersion = source["schemaVersion"];
	        this.version = source["version"];
	        this.viewport = this.convertValues(source["viewport"], CanvasViewportDTO);
	        this.theme = source["theme"];
	        this.grid = this.convertValues(source["grid"], CanvasGridDTO);
	        this.nodes = this.convertValues(source["nodes"], GraphNodeDTO);
	        this.edges = this.convertValues(source["edges"], GraphEdgeDTO);
	        this.frames = this.convertValues(source["frames"], ProductionFrameDTO);
	        this.referenceGroups = this.convertValues(source["referenceGroups"], ReferenceGroupDTO);
	        this.selectedNodeIds = source["selectedNodeIds"];
	        this.updatedAt = source["updatedAt"];
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
	export class GraphViewResult {
	    ok: boolean;
	    canvas?: ProjectCanvasDTO;
	    health?: HealthReport;
	    error?: OperationError;
	    errors: OperationError[];
	    events: ProjectEvent[];

	    static createFrom(source: any = {}) {
	        return new GraphViewResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.canvas = this.convertValues(source["canvas"], ProjectCanvasDTO);
	        this.health = this.convertValues(source["health"], HealthReport);
	        this.error = this.convertValues(source["error"], OperationError);
	        this.errors = this.convertValues(source["errors"], OperationError);
	        this.events = this.convertValues(source["events"], ProjectEvent);
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


	export class ListShotCandidatesCommand {
	    root: string;
	    scriptId?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new ListShotCandidatesCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class LoadScriptDocumentCommand {
	    root: string;
	    scriptId?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new LoadScriptDocumentCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class MarkShotContextDirtyCommand {
	    root: string;
	    shotId: string;
	    reason?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new MarkShotContextDirtyCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.shotId = source["shotId"];
	        this.reason = source["reason"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class OpenProjectCommand {
	    root: string;
	    takeover: boolean;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new OpenProjectCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.takeover = source["takeover"];
	        this.correlationId = source["correlationId"];
	    }
	}

	export class ProjectSummary {
	    projectId: string;
	    name: string;
	    type: string;
	    schemaVersion: string;
	    rootName: string;
	    openMode: string;
	    lockState: string;
	    lastCleanShutdown: boolean;
	    graphVersion: number;
	    updatedAt: string;
	    capabilities: string[];

	    static createFrom(source: any = {}) {
	        return new ProjectSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.schemaVersion = source["schemaVersion"];
	        this.rootName = source["rootName"];
	        this.openMode = source["openMode"];
	        this.lockState = source["lockState"];
	        this.lastCleanShutdown = source["lastCleanShutdown"];
	        this.graphVersion = source["graphVersion"];
	        this.updatedAt = source["updatedAt"];
	        this.capabilities = source["capabilities"];
	    }
	}
	export class OperationResult {
	    ok: boolean;
	    summary?: ProjectSummary;
	    health?: HealthReport;
	    error?: OperationError;
	    events: ProjectEvent[];

	    static createFrom(source: any = {}) {
	        return new OperationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.summary = this.convertValues(source["summary"], ProjectSummary);
	        this.health = this.convertValues(source["health"], HealthReport);
	        this.error = this.convertValues(source["error"], OperationError);
	        this.events = this.convertValues(source["events"], ProjectEvent);
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




	export class PromoteShotContextCommand {
	    root: string;
	    shotId: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new PromoteShotContextCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.shotId = source["shotId"];
	        this.correlationId = source["correlationId"];
	    }
	}

	export class RejectShotCandidateCommand {
	    root: string;
	    scriptId?: string;
	    candidateId: string;
	    rejectionReason?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new RejectShotCandidateCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.candidateId = source["candidateId"];
	        this.rejectionReason = source["rejectionReason"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class SaveGraphLayoutCommand {
	    root: string;
	    expectedGraphVersion: number;
	    viewport: CanvasViewportDTO;
	    theme: string;
	    grid: CanvasGridDTO;
	    nodes: GraphNodeLayoutCommand[];
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new SaveGraphLayoutCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.expectedGraphVersion = source["expectedGraphVersion"];
	        this.viewport = this.convertValues(source["viewport"], CanvasViewportDTO);
	        this.theme = source["theme"];
	        this.grid = this.convertValues(source["grid"], CanvasGridDTO);
	        this.nodes = this.convertValues(source["nodes"], GraphNodeLayoutCommand);
	        this.correlationId = source["correlationId"];
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
	export class SaveProjectCommand {
	    root: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new SaveProjectCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class SaveScriptDocumentCommand {
	    root: string;
	    scriptId?: string;
	    title: string;
	    sourceAssetId?: string;
	    rawText: string;
	    logline?: string;
	    synopsis?: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new SaveScriptDocumentCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.title = source["title"];
	        this.sourceAssetId = source["sourceAssetId"];
	        this.rawText = source["rawText"];
	        this.logline = source["logline"];
	        this.synopsis = source["synopsis"];
	        this.correlationId = source["correlationId"];
	    }
	}
	export class ShotCharacterRefDTO {
	    characterId?: string;
	    name: string;
	    description?: string;
	    referenceAssetId?: string;

	    static createFrom(source: any = {}) {
	        return new ShotCharacterRefDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.referenceAssetId = source["referenceAssetId"];
	    }
	}
	export class SaveShotCandidateCommand {
	    root: string;
	    scriptId?: string;
	    candidateId?: string;
	    scriptSceneId: string;
	    index: number;
	    durationSeconds: number;
	    visualDescription: string;
	    characterRefs: ShotCharacterRefDTO[];
	    sourceRange: ScriptSourceRange;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new SaveShotCandidateCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.scriptId = source["scriptId"];
	        this.candidateId = source["candidateId"];
	        this.scriptSceneId = source["scriptSceneId"];
	        this.index = source["index"];
	        this.durationSeconds = source["durationSeconds"];
	        this.visualDescription = source["visualDescription"];
	        this.characterRefs = this.convertValues(source["characterRefs"], ShotCharacterRefDTO);
	        this.sourceRange = this.convertValues(source["sourceRange"], ScriptSourceRange);
	        this.correlationId = source["correlationId"];
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
	export class ScriptSceneDTO {
	    id: string;
	    index: number;
	    title: string;
	    location: string;
	    timeOfDay: string;
	    characters: string[];
	    props: string[];
	    action: string;
	    dialogue: DialogueLineDTO[];
	    emotionalBeat: string;
	    sourceRange?: ScriptSourceRange;

	    static createFrom(source: any = {}) {
	        return new ScriptSceneDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.index = source["index"];
	        this.title = source["title"];
	        this.location = source["location"];
	        this.timeOfDay = source["timeOfDay"];
	        this.characters = source["characters"];
	        this.props = source["props"];
	        this.action = source["action"];
	        this.dialogue = this.convertValues(source["dialogue"], DialogueLineDTO);
	        this.emotionalBeat = source["emotionalBeat"];
	        this.sourceRange = this.convertValues(source["sourceRange"], ScriptSourceRange);
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
	export class ScriptDocumentDTO {
	    id: string;
	    projectId: string;
	    title: string;
	    sourceAssetId?: string;
	    rawText: string;
	    logline?: string;
	    synopsis?: string;
	    scenes: ScriptSceneDTO[];
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new ScriptDocumentDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.title = source["title"];
	        this.sourceAssetId = source["sourceAssetId"];
	        this.rawText = source["rawText"];
	        this.logline = source["logline"];
	        this.synopsis = source["synopsis"];
	        this.scenes = this.convertValues(source["scenes"], ScriptSceneDTO);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	export class ScriptDocumentResult {
	    ok: boolean;
	    document?: ScriptDocumentDTO;
	    error?: OperationError;
	    events: ProjectEvent[];

	    static createFrom(source: any = {}) {
	        return new ScriptDocumentResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.document = this.convertValues(source["document"], ScriptDocumentDTO);
	        this.error = this.convertValues(source["error"], OperationError);
	        this.events = this.convertValues(source["events"], ProjectEvent);
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
	export class ShotCardDTO {
	    id: string;
	    projectId: string;
	    sceneId: string;
	    sceneProfileId?: string;
	    sceneIdRefs?: string[];
	    sourceCandidateId: string;
	    scriptSceneId: string;
	    index: number;
	    title: string;
	    description: string;
	    durationSeconds: number;
	    aspectRatio: string;
	    shotType: string;
	    cameraMovement: string;
	    action: string;
	    emotion: string;
	    emptySceneReason?: string;
	    characterIds: string[];
	    characterRefs: ShotCharacterRefDTO[];
	    propIds?: string[];
	    referenceAssetIds?: string[];
	    sourceRange: ScriptSourceRange;
	    status: string;
	    confirmedBy: string;
	    confirmedAt: string;
	    overwrittenFields: string[];
	    packageIds?: string[];
	    resultIds?: string[];
	    continuityRuleIds?: string[];
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new ShotCardDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.sceneId = source["sceneId"];
	        this.sceneProfileId = source["sceneProfileId"];
	        this.sceneIdRefs = source["sceneIdRefs"];
	        this.sourceCandidateId = source["sourceCandidateId"];
	        this.scriptSceneId = source["scriptSceneId"];
	        this.index = source["index"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.durationSeconds = source["durationSeconds"];
	        this.aspectRatio = source["aspectRatio"];
	        this.shotType = source["shotType"];
	        this.cameraMovement = source["cameraMovement"];
	        this.action = source["action"];
	        this.emotion = source["emotion"];
	        this.emptySceneReason = source["emptySceneReason"];
	        this.characterIds = source["characterIds"];
	        this.characterRefs = this.convertValues(source["characterRefs"], ShotCharacterRefDTO);
	        this.propIds = source["propIds"];
	        this.referenceAssetIds = source["referenceAssetIds"];
	        this.sourceRange = this.convertValues(source["sourceRange"], ScriptSourceRange);
	        this.status = source["status"];
	        this.confirmedBy = source["confirmedBy"];
	        this.confirmedAt = source["confirmedAt"];
	        this.overwrittenFields = source["overwrittenFields"];
	        this.packageIds = source["packageIds"];
	        this.resultIds = source["resultIds"];
	        this.continuityRuleIds = source["continuityRuleIds"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	export class ShotCandidateDTO {
	    id: string;
	    projectId: string;
	    scriptId: string;
	    scriptSceneId: string;
	    index: number;
	    durationSeconds: number;
	    visualDescription: string;
	    characterRefs: ShotCharacterRefDTO[];
	    sourceRange: ScriptSourceRange;
	    status: string;
	    shotId?: string;
	    rejectionReason?: string;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new ShotCandidateDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.scriptId = source["scriptId"];
	        this.scriptSceneId = source["scriptSceneId"];
	        this.index = source["index"];
	        this.durationSeconds = source["durationSeconds"];
	        this.visualDescription = source["visualDescription"];
	        this.characterRefs = this.convertValues(source["characterRefs"], ShotCharacterRefDTO);
	        this.sourceRange = this.convertValues(source["sourceRange"], ScriptSourceRange);
	        this.status = source["status"];
	        this.shotId = source["shotId"];
	        this.rejectionReason = source["rejectionReason"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	export class ScriptSceneCandidateResult {
	    ok: boolean;
	    document?: ScriptDocumentDTO;
	    candidate?: ShotCandidateDTO;
	    candidates: ShotCandidateDTO[];
	    shot?: ShotCardDTO;
	    error?: OperationError;
	    events: ProjectEvent[];

	    static createFrom(source: any = {}) {
	        return new ScriptSceneCandidateResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.document = this.convertValues(source["document"], ScriptDocumentDTO);
	        this.candidate = this.convertValues(source["candidate"], ShotCandidateDTO);
	        this.candidates = this.convertValues(source["candidates"], ShotCandidateDTO);
	        this.shot = this.convertValues(source["shot"], ShotCardDTO);
	        this.error = this.convertValues(source["error"], OperationError);
	        this.events = this.convertValues(source["events"], ProjectEvent);
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





	export class ShotContextIssueDTO {
	    code: string;
	    severity: string;
	    field: string;
	    referenceId?: string;
	    userMessage: string;
	    recoveryActions: string[];

	    static createFrom(source: any = {}) {
	        return new ShotContextIssueDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.severity = source["severity"];
	        this.field = source["field"];
	        this.referenceId = source["referenceId"];
	        this.userMessage = source["userMessage"];
	        this.recoveryActions = source["recoveryActions"];
	    }
	}
	export class ShotReferenceDTO {
	    field: string;
	    kind: string;
	    referenceId: string;
	    status: string;
	    path?: string;

	    static createFrom(source: any = {}) {
	        return new ShotReferenceDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.kind = source["kind"];
	        this.referenceId = source["referenceId"];
	        this.status = source["status"];
	        this.path = source["path"];
	    }
	}
	export class ShotContextReportDTO {
	    shotId: string;
	    status: string;
	    canEnterContextReady: boolean;
	    missingFields: string[];
	    blocking: ShotContextIssueDTO[];
	    warnings: ShotContextIssueDTO[];
	    references: ShotReferenceDTO[];
	    checkedAt: string;

	    static createFrom(source: any = {}) {
	        return new ShotContextReportDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shotId = source["shotId"];
	        this.status = source["status"];
	        this.canEnterContextReady = source["canEnterContextReady"];
	        this.missingFields = source["missingFields"];
	        this.blocking = this.convertValues(source["blocking"], ShotContextIssueDTO);
	        this.warnings = this.convertValues(source["warnings"], ShotContextIssueDTO);
	        this.references = this.convertValues(source["references"], ShotReferenceDTO);
	        this.checkedAt = source["checkedAt"];
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
	export class ShotContextResult {
	    ok: boolean;
	    shot?: ShotCardDTO;
	    report: ShotContextReportDTO;
	    error?: OperationError;
	    events: ProjectEvent[];

	    static createFrom(source: any = {}) {
	        return new ShotContextResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.shot = this.convertValues(source["shot"], ShotCardDTO);
	        this.report = this.convertValues(source["report"], ShotContextReportDTO);
	        this.error = this.convertValues(source["error"], OperationError);
	        this.events = this.convertValues(source["events"], ProjectEvent);
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

	export class ValidateShotContextCommand {
	    root: string;
	    shotId: string;
	    correlationId: string;

	    static createFrom(source: any = {}) {
	        return new ValidateShotContextCommand(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.shotId = source["shotId"];
	        this.correlationId = source["correlationId"];
	    }
	}

}

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
