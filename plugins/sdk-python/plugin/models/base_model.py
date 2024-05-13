import pprint

import typing

from plugin import util

T = typing.TypeVar('T')


class Model:
    # openapiTypes: The key is attribute name and the
    # value is attribute type.
    openapi_types: typing.Dict[str, type] = {}

    # attributeMap: The key is attribute name and the
    # value is json key in definition.
    attribute_map: typing.Dict[str, str] = {}

    @classmethod
    def from_dict(cls: typing.Type[T], dikt) -> T:
        """Returns the dict as a model"""
        return util.deserialize_model(dikt, cls)

    def to_dict(self, include_none=True):
        """Returns the model properties as a dict

        :rtype: dict
        """
        result = {}

        for attr in self.openapi_types:
            value = getattr(self, attr)
            insert_attr = self.attribute_map[attr]
            if isinstance(value, list):
                result[insert_attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[insert_attr] = value.to_dict()
            elif isinstance(value, dict):
                result[insert_attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[insert_attr] = value

            if (not include_none) and (result[insert_attr] is None):
                del result[insert_attr]

        return result

    def to_json(self):
        """Returns the JSON representation of the model.
        Does not include None attrs to the dict.

        :rtype: dict
        """
        return self.to_dict(include_none=False)

    def to_str(self):
        """Returns the string representation of the model

        :rtype: str
        """
        return pprint.pformat(self.to_dict())

    def __repr__(self):
        """For `print` and `pprint`"""
        return self.to_str()

    def __eq__(self, other):
        """Returns true if both objects are equal"""
        return self.__dict__ == other.__dict__

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        return not self == other
