ace.define("ace/snippets/lsl",["require","exports","module"], function(require, exports, module) {
"use strict";

exports.snippetText = "snippet @\n\
	@${1:label};\n\
snippet CAMERA_ACTIVE\n\
	CAMERA_ACTIVE, ${1:integer isActive}, $0\n\
snippet CAMERA_BEHINDNESS_ANGLE\n\
	CAMERA_BEHINDNESS_ANGLE, ${1:float degrees}, $0\n\
snippet CAMERA_BEHINDNESS_LAG\n\
	CAMERA_BEHINDNESS_LAG, ${1:float seconds}, $0\n\
snippet CAMERA_DISTANCE\n\
	CAMERA_DISTANCE, ${1:float meters}, $0\n\
snippet CAMERA_FOCUS\n\
	CAMERA_FOCUS, ${1:vector position}, $0\n\
snippet CAMERA_FOCUS_LAG\n\
	CAMERA_FOCUS_LAG, ${1:float seconds}, $0\n\
snippet CAMERA_FOCUS_LOCKED\n\
	CAMERA_FOCUS_LOCKED, ${1:integer isLocked}, $0\n\
snippet CAMERA_FOCUS_OFFSET\n\
	CAMERA_FOCUS_OFFSET, ${1:vector meters}, $0\n\
snippet CAMERA_FOCUS_THRESHOLD\n\
	CAMERA_FOCUS_THRESHOLD, ${1:float meters}, $0\n\
snippet CAMERA_PITCH\n\
	CAMERA_PITCH, ${1:float degrees}, $0\n\
snippet CAMERA_POSITION\n\
	CAMERA_POSITION, ${1:vector position}, $0\n\
snippet CAMERA_POSITION_LAG\n\
	CAMERA_POSITION_LAG, ${1:float seconds}, $0\n\
snippet CAMERA_POSITION_LOCKED\n\
	CAMERA_POSITION_LOCKED, ${1:integer isLocked}, $0\n\
snippet CAMERA_POSITION_THRESHOLD\n\
	CAMERA_POSITION_THRESHOLD, ${1:float meters}, $0\n\
snippet CHARACTER_AVOIDANCE_MODE\n\
	CHARACTER_AVOIDANCE_MODE, ${1:integer flags}, $0\n\
snippet CHARACTER_DESIRED_SPEED\n\
	CHARACTER_DESIRED_SPEED, ${1:float speed}, $0\n\
snippet CHARACTER_DESIRED_TURN_SPEED\n\
	CHARACTER_DESIRED_TURN_SPEED, ${1:float speed}, $0\n\
snippet CHARACTER_LENGTH\n\
	CHARACTER_LENGTH, ${1:float length}, $0\n\
snippet CHARACTER_MAX_TURN_RADIUS\n\
	CHARACTER_MAX_TURN_RADIUS, ${1:float radius}, $0\n\
snippet CHARACTER_ORIENTATION\n\
	CHARACTER_ORIENTATION, ${1:integer orientation}, $0\n\
snippet CHARACTER_RADIUS\n\
	CHARACTER_RADIUS, ${1:float radius}, $0\n\
snippet CHARACTER_STAY_WITHIN_PARCEL\n\
	CHARACTER_STAY_WITHIN_PARCEL, ${1:boolean stay}, $0\n\
snippet CHARACTER_TYPE\n\
	CHARACTER_TYPE, ${1:integer type}, $0\n\
snippet HTTP_BODY_MAXLENGTH\n\
	HTTP_BODY_MAXLENGTH, ${1:integer length}, $0\n\
snippet HTTP_CUSTOM_HEADER\n\
	HTTP_CUSTOM_HEADER, ${1:string name}, ${2:string value}, $0\n\
snippet HTTP_METHOD\n\
	HTTP_METHOD, ${1:string method}, $0\n\
snippet HTTP_MIMETYPE\n\
	HTTP_MIMETYPE, ${1:string mimeType}, $0\n\
snippet HTTP_PRAGMA_NO_CACHE\n\
	HTTP_PRAGMA_NO_CACHE, ${1:integer send_header}, $0\n\
snippet HTTP_VERBOSE_THROTTLE\n\
	HTTP_VERBOSE_THROTTLE, ${1:integer noisy}, $0\n\
snippet HTTP_VERIFY_CERT\n\
	HTTP_VERIFY_CERT, ${1:integer verify}, $0\n\
snippet RC_DATA_FLAGS\n\
	RC_DATA_FLAGS, ${1:integer flags}, $0\n\
snippet RC_DETECT_PHANTOM\n\
	RC_DETECT_PHANTOM, ${1:integer dectedPhantom}, $0\n\
snippet RC_MAX_HITS\n\
	RC_MAX_HITS, ${1:integer maxHits}, $0\n\
snippet RC_REJECT_TYPES\n\
	RC_REJECT_TYPES, ${1:integer filterMask}, $0\n\
snippet at_rot_target\n\
	at_rot_target(${1:integer handle}, ${2:rotation targetrot}, ${3:rotation ourrot})\n\
	{\n\
		$0\n\
	}\n\
snippet at_target\n\
	at_target(${1:integer tnum}, ${2:vector targetpos}, ${3:vector ourpos})\n\
	{\n\
		$0\n\
	}\n\
snippet attach\n\
	attach(${1:key id})\n\
	{\n\
		$0\n\
	}\n\
snippet changed\n\
	changed(${1:integer change})\n\
	{\n\
		$0\n\
	}\n\
snippet collision\n\
	collision(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet collision_end\n\
	collision_end(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet collision_start\n\
	collision_start(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet control\n\
	control(${1:key id}, ${2:integer level}, ${3:integer edge})\n\
	{\n\
		$0\n\
	}\n\
snippet dataserver\n\
	dataserver(${1:key query_id}, ${2:string data})\n\
	{\n\
		$0\n\
	}\n\
snippet do\n\
	do\n\
	{\n\
		$0\n\
	}\n\
	while (${1:condition});\n\
snippet else\n\
	else\n\
	{\n\
		$0\n\
	}\n\
snippet email\n\
	email(${1:string time}, ${2:string address}, ${3:string subject}, ${4:string message}, ${5:integer num_left})\n\
	{\n\
		$0\n\
	}\n\
snippet experience_permissions\n\
	experience_permissions(${1:key agent_id})\n\
	{\n\
		$0\n\
	}\n\
snippet experience_permissions_denied\n\
	experience_permissions_denied(${1:key agent_id}, ${2:integer reason})\n\
	{\n\
		$0\n\
	}\n\
snippet for\n\
	for (${1:start}; ${3:condition}; ${3:step})\n\
	{\n\
		$0\n\
	}\n\
snippet http_request\n\
	http_request(${1:key request_id}, ${2:string method}, ${3:string body})\n\
	{\n\
		$0\n\
	}\n\
snippet http_response\n\
	http_response(${1:key request_id}, ${2:integer status}, ${3:list metadata}, ${4:string body})\n\
	{\n\
		$0\n\
	}\n\
snippet if\n\
	if (${1:condition})\n\
	{\n\
		$0\n\
	}\n\
snippet jump\n\
	jump ${1:label};\n\
snippet land_collision\n\
	land_collision(${1:vector pos})\n\
	{\n\
		$0\n\
	}\n\
snippet land_collision_end\n\
	land_collision_end(${1:vector pos})\n\
	{\n\
		$0\n\
	}\n\
snippet land_collision_start\n\
	land_collision_start(${1:vector pos})\n\
	{\n\
		$0\n\
	}\n\
snippet link_message\n\
	link_message(${1:integer sender_num}, ${2:integer num}, ${3:string str}, ${4:key id})\n\
	{\n\
		$0\n\
	}\n\
snippet listen\n\
	listen(${1:integer channel}, ${2:string name}, ${3:key id}, ${4:string message})\n\
	{\n\
		$0\n\
	}\n\
snippet llAbs\n\
	llAbs(${1:integer val})\n\
snippet llAcos\n\
	llAcos(${1:float val})\n\
snippet llAddToLandBanList\n\
	llAddToLandBanList(${1:key agent}, ${2:float hours});\n\
	$0\n\
snippet llAddToLandPassList\n\
	llAddToLandPassList(${1:key agent}, ${2:float hours});\n\
	$0\n\
snippet llAdjustSoundVolume\n\
	llAdjustSoundVolume(${1:float volume});\n\
	$0\n\
snippet llAgentInExperience\n\
	llAgentInExperience(${1:key agent})\n\
snippet llAllowInventoryDrop\n\
	llAllowInventoryDrop(${1:integer add});\n\
	$0\n\
snippet llAngleBetween\n\
	llAngleBetween(${1:rotation a}, ${2:rotation b})\n\
snippet llApplyImpulse\n\
	llApplyImpulse(${1:vector force}, ${2:integer local});\n\
	$0\n\
snippet llApplyRotationalImpulse\n\
	llApplyRotationalImpulse(${1:vector force}, ${2:integer local});\n\
	$0\n\
snippet llAsin\n\
	llAsin(${1:float val})\n\
snippet llAtan2\n\
	llAtan2(${1:float y}, ${2:float x})\n\
snippet llAttachToAvatar\n\
	llAttachToAvatar(${1:integer attach_point});\n\
	$0\n\
snippet llAttachToAvatarTemp\n\
	llAttachToAvatarTemp(${1:integer attach_point});\n\
	$0\n\
snippet llAvatarOnLinkSitTarget\n\
	llAvatarOnLinkSitTarget(${1:integer link})\n\
snippet llAvatarOnSitTarget\n\
	llAvatarOnSitTarget()\n\
snippet llAxes2Rot\n\
	llAxes2Rot(${1:vector fwd}, ${2:vector left}, ${3:vector up})\n\
snippet llAxisAngle2Rot\n\
	llAxisAngle2Rot(${1:vector axis}, ${2:float angle})\n\
snippet llBase64ToInteger\n\
	llBase64ToInteger(${1:string str})\n\
snippet llBase64ToString\n\
	llBase64ToString(${1:string str})\n\
snippet llBreakAllLinks\n\
	llBreakAllLinks();\n\
	$0\n\
snippet llBreakLink\n\
	llBreakLink(${1:integer link});\n\
	$0\n\
snippet llCastRay\n\
	llCastRay(${1:vector start}, ${2:vector end}, ${3:list options});\n\
	$0\n\
snippet llCeil\n\
	llCeil(${1:float val})\n\
snippet llClearCameraParams\n\
	llClearCameraParams();\n\
	$0\n\
snippet llClearLinkMedia\n\
	llClearLinkMedia(${1:integer link}, ${2:integer face});\n\
	$0\n\
snippet llClearPrimMedia\n\
	llClearPrimMedia(${1:integer face});\n\
	$0\n\
snippet llCloseRemoteDataChannel\n\
	llCloseRemoteDataChannel(${1:key channel});\n\
	$0\n\
snippet llCollisionFilter\n\
	llCollisionFilter(${1:string name}, ${2:key id}, ${3:integer accept});\n\
	$0\n\
snippet llCollisionSound\n\
	llCollisionSound(${1:string impact_sound}, ${2:float impact_volume});\n\
	$0\n\
snippet llCos\n\
	llCos(${1:float theta})\n\
snippet llCreateCharacter\n\
	llCreateCharacter(${1:list options});\n\
	$0\n\
snippet llCreateKeyValue\n\
	llCreateKeyValue(${1:string k})\n\
snippet llCreateLink\n\
	llCreateLink(${1:key target}, ${2:integer parent});\n\
	$0\n\
snippet llCSV2List\n\
	llCSV2List(${1:string src})\n\
snippet llDataSizeKeyValue\n\
	llDataSizeKeyValue()\n\
snippet llDeleteCharacter\n\
	llDeleteCharacter();\n\
	$0\n\
snippet llDeleteKeyValue\n\
	llDeleteKeyValue(${1:string k})\n\
snippet llDeleteSubList\n\
	llDeleteSubList(${1:list src}, ${2:integer start}, ${3:integer end})\n\
snippet llDeleteSubString\n\
	llDeleteSubString(${1:string src}, ${2:integer start}, ${3:integer end})\n\
snippet llDetachFromAvatar\n\
	llDetachFromAvatar();\n\
	$0\n\
snippet llDetectedGrab\n\
	llDetectedGrab(${1:integer number})\n\
snippet llDetectedGroup\n\
	llDetectedGroup(${1:integer number})\n\
snippet llDetectedKey\n\
	llDetectedKey(${1:integer number})\n\
snippet llDetectedLinkNumber\n\
	llDetectedLinkNumber(${1:integer number})\n\
snippet llDetectedName\n\
	llDetectedName(${1:integer number})\n\
snippet llDetectedOwner\n\
	llDetectedOwner(${1:integer number})\n\
snippet llDetectedPos\n\
	llDetectedPosl(${1:integer number})\n\
snippet llDetectedRot\n\
	llDetectedRot(${1:integer number})\n\
snippet llDetectedTouchBinormal\n\
	llDetectedTouchBinormal(${1:integer number})\n\
snippet llDetectedTouchFace\n\
	llDetectedTouchFace(${1:integer number})\n\
snippet llDetectedTouchNormal\n\
	llDetectedTouchNormal(${1:integer number})\n\
snippet llDetectedTouchPos\n\
	llDetectedTouchPos(${1:integer number})\n\
snippet llDetectedTouchST\n\
	llDetectedTouchST(${1:integer number})\n\
snippet llDetectedTouchUV\n\
	llDetectedTouchUV(${1:integer number})\n\
snippet llDetectedType\n\
	llDetectedType(${1:integer number})\n\
snippet llDetectedVel\n\
	llDetectedVel(${1:integer number})\n\
snippet llDialog\n\
	llDialog(${1:key agent}, ${2:string message}, ${3:list buttons}, ${4:integer channel});\n\
	$0\n\
snippet llDie\n\
	llDie();\n\
	$0\n\
snippet llDumpList2String\n\
	llDumpList2String(${1:list src}, ${2:string separator})\n\
snippet llEdgeOfWorld\n\
	llEdgeOfWorld(${1:vector pos}, ${2:vector dir})\n\
snippet llEjectFromLand\n\
	llEjectFromLand(${1:key agent});\n\
	$0\n\
snippet llEmail\n\
	llEmail(${1:string address}, ${2:string subject}, ${3:string message});\n\
	$0\n\
snippet llEscapeURL\n\
	llEscapeURL(${1:string url})\n\
snippet llEuler2Rot\n\
	llEuler2Rot(${1:vector v})\n\
snippet llExecCharacterCmd\n\
	llExecCharacterCmd(${1:integer command}, ${2:list options});\n\
	$0\n\
snippet llEvade\n\
	llEvade(${1:key target}, ${2:list options});\n\
	$0\n\
snippet llFabs\n\
	llFabs(${1:float val})\n\
snippet llFleeFrom\n\
	llFleeFrom(${1:vector position}, ${2:float distance}, ${3:list options});\n\
	$0\n\
snippet llFloor\n\
	llFloor(${1:float val})\n\
snippet llForceMouselook\n\
	llForceMouselook(${1:integer mouselook});\n\
	$0\n\
snippet llFrand\n\
	llFrand(${1:float mag})\n\
snippet llGenerateKey\n\
	llGenerateKey()\n\
snippet llGetAccel\n\
	llGetAccel()\n\
snippet llGetAgentInfo\n\
	llGetAgentInfo(${1:key id})\n\
snippet llGetAgentLanguage\n\
	llGetAgentLanguage(${1:key agent})\n\
snippet llGetAgentList\n\
	llGetAgentList(${1:integer scope}, ${2:list options})\n\
snippet llGetAgentSize\n\
	llGetAgentSize(${1:key agent})\n\
snippet llGetAlpha\n\
	llGetAlpha(${1:integer face})\n\
snippet llGetAndResetTime\n\
	llGetAndResetTime()\n\
snippet llGetAnimation\n\
	llGetAnimation(${1:key id})\n\
snippet llGetAnimationList\n\
	llGetAnimationList(${1:key agent})\n\
snippet llGetAnimationOverride\n\
	llGetAnimationOverride(${1:string anim_state})\n\
snippet llGetAttached\n\
	llGetAttached()\n\
snippet llGetAttachedList\n\
	llGetAttachedList(${1:key id})\n\
snippet llGetBoundingBox\n\
	llGetBoundingBox(${1:key object})\n\
snippet llGetCameraPos\n\
	llGetCameraPos()\n\
snippet llGetCameraRot\n\
	llGetCameraRot()\n\
snippet llGetCenterOfMass\n\
	llGetCenterOfMass()\n\
snippet llGetClosestNavPoint\n\
	llGetClosestNavPoint(${1:vector point}, ${2:list options})\n\
snippet llGetColor\n\
	llGetColor(${1:integer face})\n\
snippet llGetCreator\n\
	llGetCreator()\n\
snippet llGetDate\n\
	llGetDate()\n\
snippet llGetDisplayName\n\
	llGetDisplayName(${1:key id})\n\
snippet llGetEnergy\n\
	llGetEnergy()\n\
snippet llGetEnv\n\
	llGetEnv(${1:string name})\n\
snippet llGetExperienceDetails\n\
	llGetExperienceDetails(${1:key experience_id})\n\
snippet llGetExperienceErrorMessage\n\
	llGetExperienceErrorMessage(${1:integer error})\n\
snippet llGetForce\n\
	llGetForce()\n\
snippet llGetFreeMemory\n\
	llGetFreeMemory()\n\
snippet llGetFreeURLs\n\
	llGetFreeURLs()\n\
snippet llGetGeometricCenter\n\
	llGetGeometricCenter()\n\
snippet llGetGMTclock\n\
	llGetGMTclock()\n\
snippet llGetHTTPHeader\n\
	llGetHTTPHeader(${1:key request_id}, ${2:string header})\n\
snippet llGetInventoryCreator\n\
	llGetInventoryCreator(${1:string item})\n\
snippet llGetInventoryKey\n\
	llGetInventoryKey(${1:string name})\n\
snippet llGetInventoryName\n\
	llGetInventoryName(${1:integer type}, ${2:integer number})\n\
snippet llGetInventoryNumber\n\
	llGetInventoryNumber(${1:integer type})\n\
snippet llGetInventoryPermMask\n\
	llGetInventoryPermMask(${1:string item}, ${2:integer mask})\n\
snippet llGetInventoryType\n\
	llGetInventoryType(${1:string name})\n\
snippet llGetKey\n\
	llGetKey()\n\
snippet llGetLandOwnerAt\n\
	llGetLandOwnerAt(${1:vector pos})\n\
snippet llGetLinkKey\n\
	llGetLinkKey(${1:integer link})\n\
snippet llGetLinkMedia\n\
	llGetLinkMedia(${1:integer link}, ${2:integer face}, ${3:list params})\n\
snippet llGetLinkName\n\
	llGetLinkName(${1:integer link})\n\
snippet llGetLinkNumber\n\
	llGetLinkNumber()\n\
snippet llGetLinkNumberOfSides\n\
	llGetLinkNumberOfSides(${1:integer link})\n\
snippet llGetLinkPrimitiveParams\n\
	llGetLinkPrimitiveParams(${1:integer link}, ${2:list params})\n\
snippet llGetListEntryType\n\
	llGetListEntryType(${1:list src}, ${2:integer index})\n\
snippet llGetListLength\n\
	llGetListLength(${1:list src})\n\
snippet llGetLocalPos\n\
	llGetLocalPos()\n\
snippet llGetLocalRot\n\
	llGetLocalRot()\n\
snippet llGetMass\n\
	llGetMass()\n\
snippet llGetMassMKS\n\
	llGetMassMKS()\n\
snippet llGetMaxScaleFactor\n\
	llGetMaxScaleFactor()\n\
snippet llGetMemoryLimit\n\
	llGetMemoryLimit()\n\
snippet llGetMinScaleFactor\n\
	llGetMinScaleFactor()\n\
snippet llGetNextEmail\n\
	llGetNextEmail(${1:string address}, ${2:string subject});\n\
	$0\n\
snippet llGetNotecardLine\n\
	llGetNotecardLine(${1:string name}, ${2:integer line})\n\
snippet llGetNumberOfNotecardLines\n\
	llGetNumberOfNotecardLines(${1:string name})\n\
snippet llGetNumberOfPrims\n\
	llGetNumberOfPrims()\n\
snippet llGetNumberOfSides\n\
	llGetNumberOfSides()\n\
snippet llGetObjectDesc\n\
	llGetObjectDesc()\n\
snippet llGetObjectDetails\n\
	llGetObjectDetails(${1:key id}, ${2:list params})\n\
snippet llGetObjectMass\n\
	llGetObjectMass(${1:key id})\n\
snippet llGetObjectName\n\
	llGetObjectName()\n\
snippet llGetObjectPermMask\n\
	llGetObjectPermMask(${1:integer mask})\n\
snippet llGetObjectPrimCount\n\
	llGetObjectPrimCount(${1:key prim})\n\
snippet llGetOmega\n\
	llGetOmega()\n\
snippet llGetOwner\n\
	llGetOwner()\n\
snippet llGetOwnerKey\n\
	llGetOwnerKey(${1:key id})\n\
snippet llGetParcelDetails\n\
	llGetParcelDetails(${1:vector pos}, ${2:list params})\n\
snippet llGetParcelFlags\n\
	llGetParcelFlags(${1:vector pos})\n\
snippet llGetParcelMaxPrims\n\
	llGetParcelMaxPrims(${1:vector pos}, ${2:integer sim_wide})\n\
snippet llGetParcelMusicURL\n\
	llGetParcelMusicURL()\n\
snippet llGetParcelPrimCount\n\
	llGetParcelPrimCount(${1:vector pos}, ${2:integer category}, ${3:integer sim_wide})\n\
snippet llGetParcelPrimOwners\n\
	llGetParcelPrimOwners(${1:vector pos})\n\
snippet llGetPermissions\n\
	llGetPermissions()\n\
snippet llGetPermissionsKey\n\
	llGetPermissionsKey()\n\
snippet llGetPhysicsMaterial\n\
	llGetPhysicsMaterial()\n\
snippet llGetPos\n\
	llGetPos()\n\
snippet llGetPrimitiveParams\n\
	llGetPrimitiveParams(${1:list params})\n\
snippet llGetPrimMediaParams\n\
	llGetPrimMediaParams(${1:integer face}, ${2:list params})\n\
snippet llGetRegionAgentCount\n\
	llGetRegionAgentCount()\n\
snippet llGetRegionCorner\n\
	llGetRegionCorner()\n\
snippet llGetRegionFlags\n\
	llGetRegionFlags()\n\
snippet llGetRegionFPS\n\
	llGetRegionFPS()\n\
snippet llGetRegionName\n\
	llGetRegionName()\n\
snippet llGetRegionTimeDilation\n\
	llGetRegionTimeDilation()\n\
snippet llGetRootPosition\n\
	llGetRootPosition()\n\
snippet llGetRootRotation\n\
	llGetRootRotation()\n\
snippet llGetRot\n\
	llGetRot()\n\
snippet llGetScale\n\
	llGetScale()\n\
snippet llGetScriptName\n\
	llGetScriptName()\n\
snippet llGetScriptState\n\
	llGetScriptState(${1:string script})\n\
snippet llGetSimStats\n\
	llGetSimStats(${1:integer stat_type})\n\
snippet llGetSimulatorHostname\n\
	llGetSimulatorHostname()\n\
snippet llGetSPMaxMemory\n\
	llGetSPMaxMemory()\n\
snippet llGetStartParameter\n\
	llGetStartParameter()\n\
snippet llGetStaticPath\n\
	llGetStaticPath(${1:vector start}, ${2:vector end}, ${3:float radius}, ${4:list params})\n\
snippet llGetStatus\n\
	llGetStatus(${1:integer status})\n\
snippet llGetSubString\n\
	llGetSubString(${1:string src}, ${2:integer start}, ${3:integer end})\n\
snippet llGetSunDirection\n\
	llGetSunDirection()\n\
snippet llGetTexture\n\
	llGetTexture(${1:integer face})\n\
snippet llGetTextureOffset\n\
	llGetTextureOffset(${1:integer face})\n\
snippet llGetTextureRot\n\
	llGetTextureRot(${1:integer face})\n\
snippet llGetTextureScale\n\
	llGetTextureScale(${1:integer face})\n\
snippet llGetTime\n\
	llGetTime()\n\
snippet llGetTimeOfDay\n\
	llGetTimeOfDay()\n\
snippet llGetTimestamp\n\
	llGetTimestamp()\n\
snippet llGetTorque\n\
	llGetTorque()\n\
snippet llGetUnixTime\n\
	llGetUnixTime()\n\
snippet llGetUsedMemory\n\
	llGetUsedMemory()\n\
snippet llGetUsername\n\
	llGetUsername(${1:key id})\n\
snippet llGetVel\n\
	llGetVel()\n\
snippet llGetWallclock\n\
	llGetWallclock()\n\
snippet llGiveInventory\n\
	llGiveInventory(${1:key destination}, ${2:string inventory});\n\
	$0\n\
snippet llGiveInventoryList\n\
	llGiveInventoryList(${1:key target}, ${2:string folder}, ${3:list inventory});\n\
	$0\n\
snippet llGiveMoney\n\
	llGiveMoney(${1:key destination}, ${2:integer amount})\n\
snippet llGround\n\
	llGround(${1:vector offset})\n\
snippet llGroundContour\n\
	llGroundContour(${1:vector offset})\n\
snippet llGroundNormal\n\
	llGroundNormal(${1:vector offset})\n\
snippet llGroundRepel\n\
	llGroundRepel(${1:float height}, ${2:integer water}, ${3:float tau});\n\
	$0\n\
snippet llGroundSlope\n\
	llGroundSlope(${1:vector offset})\n\
snippet llHTTPRequest\n\
	llHTTPRequest(${1:string url}, ${2:list parameters}, ${3:string body})\n\
snippet llHTTPResponse\n\
	llHTTPResponse(${1:key request_id}, ${2:integer status}, ${3:string body});\n\
	$0\n\
snippet llInsertString\n\
	llInsertString(${1:string dst}, ${2:integer pos}, ${3:string src})\n\
snippet llInstantMessage\n\
	llInstantMessage(${1:key user}, ${2:string message});\n\
	$0\n\
snippet llIntegerToBase64\n\
	llIntegerToBase64(${1:integer number})\n\
snippet llJson2List\n\
	llJson2List(${1:string json})\n\
snippet llJsonGetValue\n\
	llJsonGetValue(${1:string json}, ${2:list specifiers})\n\
snippet llJsonSetValue\n\
	llJsonSetValue(${1:string json}, ${2:list specifiers}, ${3:string newValue})\n\
snippet llJsonValueType\n\
	llJsonValueType(${1:string json}, ${2:list specifiers})\n\
snippet llKey2Name\n\
	llKey2Name(${1:key id})\n\
snippet llKeyCountKeyValue\n\
	llKeyCountKeyValue()\n\
snippet llKeysKeyValue\n\
	llKeysKeyValue(${1:integer first}, ${2:integer count})\n\
snippet llLinkParticleSystem\n\
	llLinkParticleSystem(${1:integer link}, ${2:list rules});\n\
	$0\n\
snippet llLinkSitTarget\n\
	llLinkSitTarget(${1:integer link}, ${2:vector offset}, ${3:rotation rot});\n\
	$0\n\
snippet llList2CSV\n\
	llList2CSV(${1:list src})\n\
snippet llList2Float\n\
	llList2Float(${1:list src}, ${2:integer index})\n\
snippet llList2Integer\n\
	llList2Integer(${1:list src}, ${2:integer index})\n\
snippet llList2Json\n\
	llList2Json(${1:string type}, ${2:list values})\n\
snippet llList2Key\n\
	llList2Key(${1:list src}, ${2:integer index})\n\
snippet llList2List\n\
	llList2List(${1:list src}, ${2:integer start}, ${3:integer end})\n\
snippet llList2ListStrided\n\
	llList2ListStrided(${1:list src}, ${2:integer start}, ${3:integer end}, ${4:integer stride})\n\
snippet llList2Rot\n\
	llList2Rot(${1:list src}, ${2:integer index})\n\
snippet llList2String\n\
	llList2String(${1:list src}, ${2:integer index})\n\
snippet llList2Vector\n\
	llList2Vector(${1:list src}, ${2:integer index})\n\
snippet llListen\n\
	llListen(${1:integer channel}, ${2:string name}, ${3:key id}, ${4:string msg})\n\
snippet llListenControl\n\
	llListenControl(${1:integer handle}, ${2:integer active});\n\
	$0\n\
snippet llListenRemove\n\
	llListenRemove(${1:integer handle});\n\
	$0\n\
snippet llListFindList\n\
	llListFindList(${1:list src}, ${2:list test})\n\
snippet llListInsertList\n\
	llListInsertList(${1:list dest}, ${2:list src}, ${3:integer start})\n\
snippet llListRandomize\n\
	llListRandomize(${1:list src}, ${2:integer stride})\n\
snippet llListReplaceList\n\
	llListReplaceList(${1:list dest}, ${2:list src}, ${3:integer start}, ${4:integer end})\n\
snippet llListSort\n\
	llListSort(${1:list src}, ${2:integer stride}, ${3:integer ascending})\n\
snippet llListStatistics\n\
	llListStatistics(${1:integer operation}, ${2:list src})\n\
snippet llLoadURL\n\
	llLoadURL(${1:key agent}, ${2:string message}, ${3:string url});\n\
	$0\n\
snippet llLog\n\
	llLog(${1:float val})\n\
snippet llLog10\n\
	llLog10(${1:float val})\n\
snippet llLookAt\n\
	llLookAt(${1:vector target}, ${2:float strength}, ${3:float damping});\n\
	$0\n\
snippet llLoopSound\n\
	llLoopSound(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llLoopSoundMaster\n\
	llLoopSoundMaster(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llLoopSoundSlave\n\
	llLoopSoundSlave(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llManageEstateAccess\n\
	llManageEstateAccess(${1:integer action}, ${2:key agent})\n\
snippet llMapDestination\n\
	llMapDestination(${1:string simname}, ${2:vector pos}, ${3:vector look_at});\n\
	$0\n\
snippet llMD5String\n\
	llMD5String(${1:string src}, ${2:integer nonce})\n\
snippet llMessageLinked\n\
	llMessageLinked(${1:integer link}, ${2:integer num}, ${3:string str}, ${4:key id});\n\
	$0\n\
snippet llMinEventDelay\n\
	llMinEventDelay(${1:float delay});\n\
	$0\n\
snippet llModifyLand\n\
	llModifyLand(${1:integer action}, ${2:integer brush});\n\
	$0\n\
snippet llModPow\n\
	llModPow(${1:integer a}, ${2:integer b}, ${3:integer c})\n\
snippet llMoveToTarget\n\
	llMoveToTarget(${1:vector target}, ${2:float tau});\n\
	$0\n\
snippet llNavigateTo\n\
	llNavigateTo(${1:vector pos}, ${2:list options});\n\
	$0\n\
snippet llOffsetTexture\n\
	llOffsetTexture(${1:float u}, ${2:float v}, ${3:integer face});\n\
	$0\n\
snippet llOpenRemoteDataChannel\n\
	llOpenRemoteDataChannel();\n\
	$0\n\
snippet llOverMyLand\n\
	llOverMyLand(${1:key id})\n\
snippet llOwnerSay\n\
	llOwnerSay(${1:string msg});\n\
	$0\n\
snippet llParcelMediaCommandList\n\
	llParcelMediaCommandList(${1:list commandList});\n\
	$0\n\
snippet llParcelMediaQuery\n\
	llParcelMediaQuery(${1:list query})\n\
snippet llParseString2List\n\
	llParseString2List(${1:string src}, ${2:list separators}, ${3:list spacers})\n\
snippet llParseStringKeepNulls\n\
	llParseStringKeepNulls(${1:string src}, ${2:list separators}, ${3:list spacers})\n\
snippet llParticleSystem\n\
	llParticleSystem(${1:list rules});\n\
	$0\n\
snippet llPassCollisions\n\
	llPassCollisions(${1:integer pass});\n\
	$0\n\
snippet llPassTouches\n\
	llPassTouches(${1:integer pass});\n\
	$0\n\
snippet llPatrolPoints\n\
	llPatrolPoints(${1:list patrolPoints}, ${2:list options});\n\
	$0\n\
snippet llPlaySound\n\
	llPlaySound(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llPlaySoundSlave\n\
	llPlaySoundSlave(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llPow\n\
	llPow(${1:float base}, ${2:float exponent})\n\
snippet llPreloadSound\n\
	llPreloadSound(${1:string sound});\n\
	$0\n\
snippet llPursue\n\
	llPursue(${1:key target}, ${2:list options});\n\
	$0\n\
snippet llPushObject\n\
	llPushObject(${1:key target}, ${2:vector impulse}, ${3:vector ang_impulse}, ${4:integer local});\n\
	$0\n\
snippet llReadKeyValue\n\
	llReadKeyValue(${1:string k})\n\
snippet llRegionSay\n\
	llRegionSay(${1:integer channel}, ${2:string msg});\n\
	$0\n\
snippet llRegionSayTo\n\
	llRegionSayTo(${1:key target}, ${2:integer channel}, ${3:string msg});\n\
	$0\n\
snippet llReleaseControls\n\
	llReleaseControls();\n\
	$0\n\
snippet llReleaseURL\n\
	llReleaseURL(${1:string url});\n\
	$0\n\
snippet llRemoteDataReply\n\
	llRemoteDataReply(${1:key channel}, ${2:key message_id}, ${3:string sdata}, ${4:integer idata});\n\
	$0\n\
snippet llRemoteLoadScriptPin\n\
	llRemoteLoadScriptPin(${1:key target}, ${2:string name}, ${3:integer pin}, ${4:integer running}, ${5:integer start_param});\n\
	$0\n\
snippet llRemoveFromLandBanList\n\
	llRemoveFromLandBanList(${1:key agent});\n\
	$0\n\
snippet llRemoveFromLandPassList\n\
	llRemoveFromLandPassList(${1:key agent});\n\
	$0\n\
snippet llRemoveInventory\n\
	llRemoveInventory(${1:string item});\n\
	$0\n\
snippet llRemoveVehicleFlags\n\
	llRemoveVehicleFlags(${1:integer flags});\n\
	$0\n\
snippet llRequestAgentData\n\
	llRequestAgentData(${1:key id}, ${2:integer data})\n\
snippet llRequestDisplayName\n\
	llRequestDisplayName(${1:key id})\n\
snippet llRequestExperiencePermissions\n\
	llRequestExperiencePermissions(${1:key agent}, ${2:string name})\n\
snippet llRequestInventoryData\n\
	llRequestInventoryData(${1:string name})\n\
snippet llRequestPermissions\n\
	llRequestPermissions(${1:key agent}, ${2:integer permissions})\n\
snippet llRequestSecureURL\n\
	llRequestSecureURL()\n\
snippet llRequestSimulatorData\n\
	llRequestSimulatorData(${1:string region}, ${2:integer data})\n\
snippet llRequestURL\n\
	llRequestURL()\n\
snippet llRequestUsername\n\
	llRequestUsername(${1:key id})\n\
snippet llResetAnimationOverride\n\
	llResetAnimationOverride(${1:string anim_state});\n\
	$0\n\
snippet llResetLandBanList\n\
	llResetLandBanList();\n\
	$0\n\
snippet llResetLandPassList\n\
	llResetLandPassList();\n\
	$0\n\
snippet llResetOtherScript\n\
	llResetOtherScript(${1:string name});\n\
	$0\n\
snippet llResetScript\n\
	llResetScript();\n\
	$0\n\
snippet llResetTime\n\
	llResetTime();\n\
	$0\n\
snippet llReturnObjectsByID\n\
	llReturnObjectsByID(${1:list objects})\n\
snippet llReturnObjectsByOwner\n\
	llReturnObjectsByOwner(${1:key owner}, ${2:integer scope})\n\
snippet llRezAtRoot\n\
	llRezAtRoot(${1:string inventory}, ${2:vector position}, ${3:vector velocity}, ${4:rotation rot}, ${5:integer param});\n\
	$0\n\
snippet llRezObject\n\
	llRezObject(${1:string inventory}, ${2:vector pos}, ${3:vector vel}, ${4:rotation rot}, ${5:integer param});\n\
	$0\n\
snippet llRot2Angle\n\
	llRot2Angle(${1:rotation rot})\n\
snippet llRot2Axis\n\
	llRot2Axis(${1:rotation rot})\n\
snippet llRot2Euler\n\
	llRot2Euler(${1:rotation quat})\n\
snippet llRot2Fwd\n\
	llRot2Fwd(${1:rotation q})\n\
snippet llRot2Left\n\
	llRot2Left(${1:rotation q})\n\
snippet llRot2Up\n\
	llRot2Up(${1:rotation q})\n\
snippet llRotateTexture\n\
	llRotateTexture(${1:float angle}, ${2:integer face});\n\
	$0\n\
snippet llRotBetween\n\
	llRotBetween(${1:vector start}, ${2:vector end})\n\
snippet llRotLookAt\n\
	llRotLookAt(${1:rotation target_direction}, ${2:float strength}, ${3:float damping});\n\
	$0\n\
snippet llRotTarget\n\
	llRotTarget(${1:rotation rot}, ${2:float error})\n\
snippet llRotTargetRemove\n\
	llRotTargetRemove(${1:integer handle});\n\
	$0\n\
snippet llRound\n\
	llRound(${1:float val})\n\
snippet llSameGroup\n\
	llSameGroup(${1:key group})\n\
snippet llSay\n\
	llSay(${1:integer channel}, ${2:string msg});\n\
	$0\n\
snippet llScaleByFactor\n\
	llScaleByFactor(${1:float scaling_factor})\n\
snippet llScaleTexture\n\
	llScaleTexture(${1:float u}, ${2:float v}, ${3:integer face});\n\
	$0\n\
snippet llScriptDanger\n\
	llScriptDanger(${1:vector pos})\n\
snippet llScriptProfiler\n\
	llScriptProfiler(${1:integer flags});\n\
	$0\n\
snippet llSendRemoteData\n\
	llSendRemoteData(${1:key channel}, ${2:string dest}, ${3:integer idata}, ${4:string sdata})\n\
snippet llSensor\n\
	llSensor(${1:string name}, ${2:key id}, ${3:integer type}, ${4:float range}, ${5:float arc});\n\
	$0\n\
snippet llSensorRepeat\n\
	llSensorRepeat(${1:string name}, ${2:key id}, ${3:integer type}, ${4:float range}, ${5:float arc}, ${6:float rate});\n\
	$0\n\
snippet llSetAlpha\n\
	llSetAlpha(${1:float alpha}, ${2:integer face});\n\
	$0\n\
snippet llSetAngularVelocity\n\
	llSetAngularVelocity(${1:vector force}, ${2:integer local});\n\
	$0\n\
snippet llSetAnimationOverride\n\
	llSetAnimationOverride(${1:string anim_state}, ${2:string anim})\n\
snippet llSetBuoyancy\n\
	llSetBuoyancy(${1:float buoyancy});\n\
	$0\n\
snippet llSetCameraAtOffset\n\
	llSetCameraAtOffset(${1:vector offset});\n\
	$0\n\
snippet llSetCameraEyeOffset\n\
	llSetCameraEyeOffset(${1:vector offset});\n\
	$0\n\
snippet llSetCameraParams\n\
	llSetCameraParams(${1:list rules});\n\
	$0\n\
snippet llSetClickAction\n\
	llSetClickAction(${1:integer action});\n\
	$0\n\
snippet llSetColor\n\
	llSetColor(${1:vector color}, ${2:integer face});\n\
	$0\n\
snippet llSetContentType\n\
	llSetContentType(${1:key request_id}, ${2:integer content_type});\n\
	$0\n\
snippet llSetDamage\n\
	llSetDamage(${1:float damage});\n\
	$0\n\
snippet llSetForce\n\
	llSetForce(${1:vector force}, ${2:integer local});\n\
	$0\n\
snippet llSetForceAndTorque\n\
	llSetForceAndTorque(${1:vector force}, ${2:vector torque}, ${3:integer local});\n\
	$0\n\
snippet llSetHoverHeight\n\
	llSetHoverHeight(${1:float height}, ${2:integer water}, ${3:float tau});\n\
	$0\n\
snippet llSetKeyframedMotion\n\
	llSetKeyframedMotion(${1:list keyframes}, ${2:list options});\n\
	$0\n\
snippet llSetLinkAlpha\n\
	llSetLinkAlpha(${1:integer link}, ${2:float alpha}, ${3:integer face});\n\
	$0\n\
snippet llSetLinkCamera\n\
	llSetLinkCamera(${1:integer link}, ${2:vector eye}, ${3:vector at});\n\
	$0\n\
snippet llSetLinkColor\n\
	llSetLinkColor(${1:integer link}, ${2:vector color}, ${3:integer face});\n\
	$0\n\
snippet llSetLinkMedia\n\
	llSetLinkMedia(${1:integer link}, ${2:integer face}, ${3:list params});\n\
	$0\n\
snippet llSetLinkPrimitiveParams\n\
	llSetLinkPrimitiveParams(${1:integer link}, ${2:list rules});\n\
	$0\n\
snippet llSetLinkPrimitiveParamsFast\n\
	llSetLinkPrimitiveParamsFast(${1:integer link}, ${2:list rules});\n\
	$0\n\
snippet llSetLinkTexture\n\
	llSetLinkTexture(${1:integer link}, ${2:string texture}, ${3:integer face});\n\
	$0\n\
snippet llSetLinkTextureAnim\n\
	llSetLinkTextureAnim(${1:integer link}, ${2:integer mode}, ${3:integer face}, ${4:integer sizex}, ${5:integer sizey}, ${6:float start}, ${7:float length}, ${8:float rate});\n\
	$0\n\
snippet llSetLocalRot\n\
	llSetLocalRot(${1:rotation rot});\n\
	$0\n\
snippet llSetMemoryLimit\n\
	llSetMemoryLimit(${1:integer limit})\n\
snippet llSetObjectDesc\n\
	llSetObjectDesc(${1:string description});\n\
	$0\n\
snippet llSetObjectName\n\
	llSetObjectName(${1:string name});\n\
	$0\n\
snippet llSetParcelMusicURL\n\
	llSetParcelMusicURL(${1:string url});\n\
	$0\n\
snippet llSetPayPrice\n\
	llSetPayPrice(${1:integer price}, [${2:integer price_button_a}, ${3:integer price_button_b}, ${4:integer price_button_c}, ${5:integer price_button_d}]);\n\
	$0\n\
snippet llSetPhysicsMaterial\n\
	llSetPhysicsMaterial(${1:integer mask}, ${2:float gravity_multiplier}, ${3:float restitution}, ${4:float friction}, ${5:float density});\n\
	$0\n\
snippet llSetPos\n\
	llSetPos(${1:vector pos});\n\
	$0\n\
snippet llSetPrimitiveParams\n\
	llSetPrimitiveParams(${1:list rules});\n\
	$0\n\
snippet llSetPrimMediaParams\n\
	llSetPrimMediaParams(${1:integer face}, ${2:list params});\n\
	$0\n\
snippet llSetRegionPos\n\
	llSetRegionPos(${1:vector position})\n\
snippet llSetRemoteScriptAccessPin\n\
	llSetRemoteScriptAccessPin(${1:integer pin});\n\
	$0\n\
snippet llSetRot\n\
	llSetRot(${1:rotation rot});\n\
	$0\n\
snippet llSetScale\n\
	llSetScale(${1:vector size});\n\
	$0\n\
snippet llSetScriptState\n\
	llSetScriptState(${1:string name}, ${2:integer run});\n\
	$0\n\
snippet llSetSitText\n\
	llSetSitText(${1:string text});\n\
	$0\n\
snippet llSetSoundQueueing\n\
	llSetSoundQueueing(${1:integer queue});\n\
	$0\n\
snippet llSetSoundRadius\n\
	llSetSoundRadius(${1:float radius});\n\
	$0\n\
snippet llSetStatus\n\
	llSetStatus(${1:integer status}, ${2:integer value});\n\
	$0\n\
snippet llSetText\n\
	llSetText(${1:string text}, ${2:vector color}, ${3:float alpha});\n\
	$0\n\
snippet llSetTexture\n\
	llSetTexture(${1:string texture}, ${2:integer face});\n\
	$0\n\
snippet llSetTextureAnim\n\
	llSetTextureAnim(${1:integer mode}, ${2:integer face}, ${3:integer sizex}, ${4:integer sizey}, ${5:float start}, ${6:float length}, ${7:float rate});\n\
	$0\n\
snippet llSetTimerEvent\n\
	llSetTimerEvent(${1:float sec});\n\
	$0\n\
snippet llSetTorque\n\
	llSetTorque(${1:vector torque}, ${2:integer local});\n\
	$0\n\
snippet llSetTouchText\n\
	llSetTouchText(${1:string text});\n\
	$0\n\
snippet llSetVehicleFlags\n\
	llSetVehicleFlags(${1:integer flags});\n\
	$0\n\
snippet llSetVehicleFloatParam\n\
	llSetVehicleFloatParam(${1:integer param}, ${2:float value});\n\
	$0\n\
snippet llSetVehicleRotationParam\n\
	llSetVehicleRotationParam(${1:integer param}, ${2:rotation rot});\n\
	$0\n\
snippet llSetVehicleType\n\
	llSetVehicleType(${1:integer type});\n\
	$0\n\
snippet llSetVehicleVectorParam\n\
	llSetVehicleVectorParam(${1:integer param}, ${2:vector vec});\n\
	$0\n\
snippet llSetVelocity\n\
	llSetVelocity(${1:vector force}, ${2:integer local});\n\
	$0\n\
snippet llSHA1String\n\
	llSHA1String(${1:string src})\n\
snippet llShout\n\
	llShout(${1:integer channel}, ${2:string msg});\n\
	$0\n\
snippet llSin\n\
	llSin(${1:float theta})\n\
snippet llSitTarget\n\
	llSitTarget(${1:vector offset}, ${2:rotation rot});\n\
	$0\n\
snippet llSleep\n\
	llSleep(${1:float sec});\n\
	$0\n\
snippet llSqrt\n\
	llSqrt(${1:float val})\n\
snippet llStartAnimation\n\
	llStartAnimation(${1:string anim});\n\
	$0\n\
snippet llStopAnimation\n\
	llStopAnimation(${1:string anim});\n\
	$0\n\
snippet llStopHover\n\
	llStopHover();\n\
	$0\n\
snippet llStopLookAt\n\
	llStopLookAt();\n\
	$0\n\
snippet llStopMoveToTarget\n\
	llStopMoveToTarget();\n\
	$0\n\
snippet llStopSound\n\
	llStopSound();\n\
	$0\n\
snippet llStringLength\n\
	llStringLength(${1:string str})\n\
snippet llStringToBase64\n\
	llStringToBase64(${1:string str})\n\
snippet llStringTrim\n\
	llStringTrim(${1:string src}, ${2:integer type})\n\
snippet llSubStringIndex\n\
	llSubStringIndex(${1:string source}, ${2:string pattern})\n\
snippet llTakeControls\n\
	llTakeControls(${1:integer controls}, ${2:integer accept}, ${3:integer pass_on});\n\
	$0\n\
snippet llTan\n\
	llTan(${1:float theta})\n\
snippet llTarget\n\
	llTarget(${1:vector position}, ${2:float range})\n\
snippet llTargetOmega\n\
	llTargetOmega(${1:vector axis}, ${2:float spinrate}, ${3:float gain});\n\
	$0\n\
snippet llTargetRemove\n\
	llTargetRemove(${1:integer handle});\n\
	$0\n\
snippet llTeleportAgent\n\
	llTeleportAgent(${1:key agent}, ${2:string landmark}, ${3:vector position}, ${4:vector look_at});\n\
	$0\n\
snippet llTeleportAgentGlobalCoords\n\
	llTeleportAgentGlobalCoords(${1:key agent}, ${2:vector global_coordinates}, ${3:vector region_coordinates}, ${4:vector look_at});\n\
	$0\n\
snippet llTeleportAgentHome\n\
	llTeleportAgentHome(${1:key agent});\n\
	$0\n\
snippet llTextBox\n\
	llTextBox(${1:key agent}, ${2:string message}, ${3:integer channel});\n\
	$0\n\
snippet llToLower\n\
	llToLower(${1:string src})\n\
snippet llToUpper\n\
	llToUpper(${1:string src})\n\
snippet llTransferLindenDollars\n\
	llTransferLindenDollars(${1:key destination}, ${2:integer amount})\n\
snippet llTriggerSound\n\
	llTriggerSound(${1:string sound}, ${2:float volume});\n\
	$0\n\
snippet llTriggerSoundLimited\n\
	llTriggerSoundLimited(${1:string sound}, ${2:float volume}, ${3:vector top_north_east}, ${4:vector bottom_south_west});\n\
	$0\n\
snippet llUnescapeURL\n\
	llUnescapeURL(${1:string url})\n\
snippet llUnSit\n\
	llUnSit(${1:key id});\n\
	$0\n\
snippet llUpdateCharacter\n\
	llUpdateCharacter(${1:list options})\n\
snippet llUpdateKeyValue\n\
	llUpdateKeyValue(${1:string k}, ${2:string v}, ${3:integer checked}, ${4:string ov})\n\
snippet llVecDist\n\
	llVecDist(${1:vector vec_a}, ${2:vector vec_b})\n\
snippet llVecMag\n\
	llVecMag(${1:vector vec})\n\
snippet llVecNorm\n\
	llVecNorm(${1:vector vec})\n\
snippet llVolumeDetect\n\
	llVolumeDetect(${1:integer detect});\n\
	$0\n\
snippet llWanderWithin\n\
	llWanderWithin(${1:vector origin}, ${2:vector dist}, ${3:list options});\n\
	$0\n\
snippet llWater\n\
	llWater(${1:vector offset});\n\
	$0\n\
snippet llWhisper\n\
	llWhisper(${1:integer channel}, ${2:string msg});\n\
	$0\n\
snippet llWind\n\
	llWind(${1:vector offset});\n\
	$0\n\
snippet llXorBase64\n\
	llXorBase64(${1:string str1}, ${2:string str2})\n\
snippet money\n\
	money(${1:key id}, ${2:integer amount})\n\
	{\n\
		$0\n\
	}\n\
snippet object_rez\n\
	object_rez(${1:key id})\n\
	{\n\
		$0\n\
	}\n\
snippet on_rez\n\
	on_rez(${1:integer start_param})\n\
	{\n\
		$0\n\
	}\n\
snippet path_update\n\
	path_update(${1:integer type}, ${2:list reserved})\n\
	{\n\
		$0\n\
	}\n\
snippet remote_data\n\
	remote_data(${1:integer event_type}, ${2:key channel}, ${3:key message_id}, ${4:string sender}, ${5:integer idata}, ${6:string sdata})\n\
	{\n\
		$0\n\
	}\n\
snippet run_time_permissions\n\
	run_time_permissions(${1:integer perm})\n\
	{\n\
		$0\n\
	}\n\
snippet sensor\n\
	sensor(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet state\n\
	state ${1:name}\n\
snippet touch\n\
	touch(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet touch_end\n\
	touch_end(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet touch_start\n\
	touch_start(${1:integer index})\n\
	{\n\
		$0\n\
	}\n\
snippet transaction_result\n\
	transaction_result(${1:key id}, ${2:integer success}, ${3:string data})\n\
	{\n\
		$0\n\
	}\n\
snippet while\n\
	while (${1:condition})\n\
	{\n\
		$0\n\
	}\n\
";
exports.scope = "lsl";

});
