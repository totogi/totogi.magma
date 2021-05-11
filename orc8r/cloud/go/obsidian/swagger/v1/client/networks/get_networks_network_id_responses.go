// Code generated by go-swagger; DO NOT EDIT.

package networks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "magma/orc8r/cloud/go/obsidian/swagger/v1/models"
)

// GetNetworksNetworkIDReader is a Reader for the GetNetworksNetworkID structure.
type GetNetworksNetworkIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetNetworksNetworkIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetNetworksNetworkIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewGetNetworksNetworkIDDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetNetworksNetworkIDOK creates a GetNetworksNetworkIDOK with default headers values
func NewGetNetworksNetworkIDOK() *GetNetworksNetworkIDOK {
	return &GetNetworksNetworkIDOK{}
}

/*GetNetworksNetworkIDOK handles this case with default header values.

Network description
*/
type GetNetworksNetworkIDOK struct {
	Payload *models.Network
}

func (o *GetNetworksNetworkIDOK) Error() string {
	return fmt.Sprintf("[GET /networks/{network_id}][%d] getNetworksNetworkIdOK  %+v", 200, o.Payload)
}

func (o *GetNetworksNetworkIDOK) GetPayload() *models.Network {
	return o.Payload
}

func (o *GetNetworksNetworkIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Network)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetNetworksNetworkIDDefault creates a GetNetworksNetworkIDDefault with default headers values
func NewGetNetworksNetworkIDDefault(code int) *GetNetworksNetworkIDDefault {
	return &GetNetworksNetworkIDDefault{
		_statusCode: code,
	}
}

/*GetNetworksNetworkIDDefault handles this case with default header values.

Unexpected Error
*/
type GetNetworksNetworkIDDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get networks network ID default response
func (o *GetNetworksNetworkIDDefault) Code() int {
	return o._statusCode
}

func (o *GetNetworksNetworkIDDefault) Error() string {
	return fmt.Sprintf("[GET /networks/{network_id}][%d] GetNetworksNetworkID default  %+v", o._statusCode, o.Payload)
}

func (o *GetNetworksNetworkIDDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetNetworksNetworkIDDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}