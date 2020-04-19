package ctrlupnp

import (
	"fmt"
	"net/http"
)

func (c *Controller) ServeDeviceXML(_ *http.Request) *Response {
	response := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
	<root xmlns="urn:schemas-upnp-org:device-1-0">
		<specVersion>
			<major>1</major>
			<minor>0</minor>
		</specVersion>
		<device>
			<UDN>uuid:%s</UDN>
			<friendlyName>My Gonic Server</friendlyName>
			<deviceType>urn:schemas-upnp-org:device:MediaServer:1</deviceType>
			<manufacturer>Stefan Richter</manufacturer>
			<manufacturerURL>http://github.com/gonic/gonic</manufacturerURL>
			<modelName>Gonic</modelName>
			<modelNumber>1</modelNumber>
			<modelURL>http://github.com/gonic/gonic</modelURL>
			<serialNumber>42</serialNumber>
			<dlna:X_DLNADOC xmlns:dlna="urn:schemas-dlna-org:device-1-0">DMS-1.50</dlna:X_DLNADOC>
			<serviceList>
				<service>
					<serviceType>urn:schemas-upnp-org:service:ConnectionManager:1</serviceType>
					<serviceId>urn:upnp-org:serviceId:ConnectionManager</serviceId>
					<SCPDURL>/upnp/cms.xml</SCPDURL>
					<controlURL>/upnp/cms_ctrl</controlURL>
					<eventSubURL>/upnp/cms_evt</eventSubURL>
				</service>
				<service>
					<serviceType>urn:schemas-upnp-org:service:ContentDirectory:1</serviceType>
					<serviceId>urn:upnp-org:serviceId:ContentDirectory</serviceId>
					<SCPDURL>/upnp/cds.xml</SCPDURL>
					<controlURL>/upnp/cds_ctrl</controlURL>
					<eventSubURL>/upnp/cds_evt</eventSubURL>
				</service>

			</serviceList>
		</device>
	</root>`, DeviceUUID)
	return &Response{code: http.StatusOK, responseData: []byte(response)}
}

func (c *Controller) ServeContentDirectoryXML(_ *http.Request) *Response {
	response := fmt.Sprintf(`<?xml version="1.0"?>
	<scpd xmlns="urn:schemas-upnp-org:service-1-0">
	  <specVersion>
		<major>1</major>
		<minor>0</minor>
	  </specVersion>
	  <actionList>
		<action>
		  <name>GetSearchCapabilities</name>
		  <argumentList>
			<argument>
			  <name>SearchCaps</name>
			  <direction>out</direction>
			  <relatedStateVariable>SearchCapabilities</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>GetSortCapabilities</name>
		  <argumentList>
			<argument>
			  <name>SortCaps</name>
			  <direction>out</direction>
			  <relatedStateVariable>SortCapabilities</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>GetSystemUpdateID</name>
		  <argumentList>
			<argument>
			  <name>Id</name>
			  <direction>out</direction>
			  <relatedStateVariable>SystemUpdateID</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>Browse</name>
		  <argumentList>
			<argument>
			  <name>ObjectID</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_ObjectID</relatedStateVariable>
			</argument>
			<argument>
			  <name>BrowseFlag</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_BrowseFlag</relatedStateVariable>
			</argument>
			<argument>
			  <name>Filter</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Filter</relatedStateVariable>
			</argument>
			<argument>
			  <name>StartingIndex</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Index</relatedStateVariable>
			</argument>
			<argument>
			  <name>RequestedCount</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>SortCriteria</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_SortCriteria</relatedStateVariable>
			</argument>
			<argument>
			  <name>Result</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Result</relatedStateVariable>
			</argument>
			<argument>
			  <name>NumberReturned</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>TotalMatches</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>UpdateID</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_UpdateID</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>Search</name>
		  <argumentList>
			<argument>
			  <name>ContainerID</name>
			  <direction>in</direction>
			<relatedStateVariable>A_ARG_TYPE_ObjectID</relatedStateVariable>
			</argument>
			<argument>
			  <name>SearchCriteria</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_SearchCriteria</relatedStateVariable>
			</argument>
			<argument>
			  <name>Filter</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Filter</relatedStateVariable>
			</argument>
			<argument>
			  <name>StartingIndex</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Index</relatedStateVariable>
			</argument>
			<argument>
			  <name>RequestedCount</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>SortCriteria</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_SortCriteria</relatedStateVariable>
			</argument>
			<argument>
			  <name>Result</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Result</relatedStateVariable>
			</argument>
			<argument>
			  <name>NumberReturned</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>TotalMatches</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Count</relatedStateVariable>
			</argument>
			<argument>
			  <name>UpdateID</name>
			  <direction>out</direction>
		   <relatedStateVariable>A_ARG_TYPE_UpdateID</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
	  </actionList>
	  <serviceStateTable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_ObjectID</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_Result</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_SearchCriteria</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_BrowseFlag</name>
		  <dataType>string</dataType>
		  <allowedValueList>
			<allowedValue>BrowseMetadata</allowedValue>
			<allowedValue>BrowseDirectChildren</allowedValue>
		  </allowedValueList>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_Filter</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_SortCriteria</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_Index</name>
		  <dataType>ui4</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_Count</name>
		  <dataType>ui4</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_UpdateID</name>
		  <dataType>ui4</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>SearchCapabilities</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>SortCapabilities</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="yes">
		  <name>SystemUpdateID</name>
		  <dataType>ui4</dataType>
		</stateVariable>
		<stateVariable sendEvents="yes">
		  <name>ContainerUpdateIDs</name>
		  <dataType>string</dataType>
		</stateVariable>
	  </serviceStateTable>
	</scpd>`)
	return &Response{code: http.StatusOK, responseData: []byte(response)}
}

func (c *Controller) ServeConnectionManagerXML(_ *http.Request) *Response {
	response := fmt.Sprintf(`<?xml version="1.0"?>
	<scpd xmlns="urn:schemas-upnp-org:service-1-0">
	  <specVersion>
		<major>1</major>
		<minor>0</minor>
	  </specVersion>
	  <actionList>
		<action>
		  <name>GetProtocolInfo</name>
		  <argumentList>
			<argument>
			  <name>Source</name>
			  <direction>out</direction>
			  <relatedStateVariable>SourceProtocolInfo</relatedStateVariable>
			</argument>
			<argument>
			  <name>Sink</name>
			  <direction>out</direction>
			  <relatedStateVariable>SinkProtocolInfo</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>GetCurrentConnectionIDs</name>
		  <argumentList>
			<argument>
			  <name>ConnectionIDs</name>
			  <direction>out</direction>
			  <relatedStateVariable>CurrentConnectionIDs</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
		<action>
		  <name>GetCurrentConnectionInfo></name>
		  <argumentList>
			<argument>
			  <name>ConnectionID</name>
			  <direction>in</direction>
			  <relatedStateVariable>A_ARG_TYPE_ConnectionID</relatedStateVariable>
			</argument>
			<argument>
			  <name>RcsID</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_RcsID</relatedStateVariable>
			</argument>
			<argument>
			  <name>AVTransportID</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_AVTransportID</relatedStateVariable>
			</argument>
			<argument>
			  <name>ProtocolInfo</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_ProtocolInfo</relatedStateVariable>
			</argument>
			<argument>
			  <name>PeerConnectionManager</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_ConnectionManager</relatedStateVariable>
			</argument>
			<argument>
			  <name>PeerConnectionID</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_ConnectionID</relatedStateVariable>
			</argument>
			<argument>
			  <name>Direction</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_Direction</relatedStateVariable>
			</argument>
			<argument>
			  <name>Status</name>
			  <direction>out</direction>
			  <relatedStateVariable>A_ARG_TYPE_ConnectionStatus</relatedStateVariable>
			</argument>
		  </argumentList>
		</action>
	  </actionList>
	  <serviceStateTable>
		<stateVariable sendEvents="yes">
		  <name>SourceProtocolInfo</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="yes">
		  <name>SinkProtocolInfo</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="yes">
		  <name>CurrentConnectionIDs</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_ConnectionStatus</name>
		  <dataType>string</dataType>
		  <allowedValueList>
			<allowedValue>OK</allowedValue>
			<allowedValue>ContentFormatMismatch</allowedValue>
			<allowedValue>InsufficientBandwidth</allowedValue>
			<allowedValue>UnreliableChannel</allowedValue>
			<allowedValue>Unknown</allowedValue>
		  </allowedValueList>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_ConnectionManager</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_Direction</name>
		  <dataType>string</dataType>
		  <allowedValueList>
			<allowedValue>Input</allowedValue>
			<allowedValue>Output</allowedValue>
		  </allowedValueList>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_ProtocolInfo</name>
		  <dataType>string</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_ConnectionID</name>
		  <dataType>i4</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_AVTransportID</name>
		  <dataType>i4</dataType>
		</stateVariable>
		<stateVariable sendEvents="no">
		  <name>A_ARG_TYPE_RcsID</name>
		  <dataType>i4</dataType>
		</stateVariable>
	  </serviceStateTable>
	</scpd>`)
	return &Response{code: http.StatusOK, responseData: []byte(response)}
}