using System.Collections.Immutable;
using System.Globalization;
using System.Reflection;
using System.Reflection.Metadata;
using System.Reflection.PortableExecutable;
using System.Text.Json;
using System.Text.Json.Serialization;
using MessagePack;

record MetadataExport(
    [property: JsonPropertyName("constants")] IReadOnlyList<MetadataConstant> Constants,
    [property: JsonPropertyName("constant_methods")] IReadOnlyList<MetadataConstantMethod> ConstantMethods,
    [property: JsonPropertyName("initializers")] IReadOnlyList<MetadataInitializer> Initializers,
    [property: JsonPropertyName("guids")] IReadOnlyList<MetadataGuid> Guids);

record MetadataConstant(
    [property: JsonPropertyName("namespace")] string Namespace,
    [property: JsonPropertyName("declaring_type")] string DeclaringType,
    [property: JsonPropertyName("name")] string Name,
    [property: JsonPropertyName("managed_type")] string ManagedType,
    [property: JsonPropertyName("native_type")] string? NativeType,
    [property: JsonPropertyName("documentation")] string? Documentation,
    [property: JsonPropertyName("comment")] string? Comment,
    [property: JsonPropertyName("kind")] string Kind,
    [property: JsonPropertyName("value")] string Value,
    [property: JsonPropertyName("enum")] bool Enum);

record MetadataConstantMethod(
    [property: JsonPropertyName("namespace")] string Namespace,
    [property: JsonPropertyName("declaring_type")] string DeclaringType,
    [property: JsonPropertyName("name")] string Name,
    [property: JsonPropertyName("return_type")] string ReturnType,
    [property: JsonPropertyName("documentation")] string? Documentation,
    [property: JsonPropertyName("comment")] string? Comment,
    [property: JsonPropertyName("value")] string Value);

[MessagePackObject]
public sealed class ApiDetails
{
    [Key(0)]
    public Uri? HelpLink { get; set; }

    [Key(1)]
    public string? Description { get; set; }

    [Key(2)]
    public string? Remarks { get; set; }

    [Key(3)]
    public Dictionary<string, string> Parameters { get; set; } = new();

    [Key(4)]
    public Dictionary<string, string> Fields { get; set; } = new();

    [Key(5)]
    public string? ReturnValue { get; set; }
}

record MetadataInitializer(
    [property: JsonPropertyName("namespace")] string Namespace,
    [property: JsonPropertyName("declaring_type")] string DeclaringType,
    [property: JsonPropertyName("name")] string Name,
    [property: JsonPropertyName("managed_type")] string ManagedType,
    [property: JsonPropertyName("value")] string Value);

record MetadataGuid(
    [property: JsonPropertyName("namespace")] string Namespace,
    [property: JsonPropertyName("name")] string Name,
    [property: JsonPropertyName("kind")] string Kind,
    [property: JsonPropertyName("value")] string Value);

sealed class TypeNameProvider :
    ISignatureTypeProvider<string, object?>,
    ICustomAttributeTypeProvider<string>
{
    public string GetArrayType(string elementType, ArrayShape shape)
    {
        return elementType + "[]";
    }

    public string GetByReferenceType(string elementType)
    {
        return elementType + "&";
    }

    public string GetFunctionPointerType(MethodSignature<string> signature)
    {
        return "fnptr";
    }

    public string GetGenericInstantiation(string genericType, ImmutableArray<string> typeArguments)
    {
        return genericType + "<" + string.Join(",", typeArguments) + ">";
    }

    public string GetGenericMethodParameter(object? genericContext, int index)
    {
        return "!!" + index.ToString(CultureInfo.InvariantCulture);
    }

    public string GetGenericTypeParameter(object? genericContext, int index)
    {
        return "!" + index.ToString(CultureInfo.InvariantCulture);
    }

    public string GetModifiedType(string modifier, string unmodifiedType, bool isRequired)
    {
        return unmodifiedType;
    }

    public string GetPinnedType(string elementType)
    {
        return elementType;
    }

    public string GetPointerType(string elementType)
    {
        return elementType + "*";
    }

    public string GetPrimitiveType(PrimitiveTypeCode typeCode)
    {
        return typeCode.ToString();
    }

    public string GetSZArrayType(string elementType)
    {
        return elementType + "[]";
    }

    public string GetTypeFromDefinition(MetadataReader metadataReader, TypeDefinitionHandle handle, byte rawTypeKind)
    {
        var definition = metadataReader.GetTypeDefinition(handle);

        return Qualify(metadataReader.GetString(definition.Namespace), metadataReader.GetString(definition.Name));
    }

    public string GetTypeFromReference(MetadataReader metadataReader, TypeReferenceHandle handle, byte rawTypeKind)
    {
        var reference = metadataReader.GetTypeReference(handle);

        return Qualify(metadataReader.GetString(reference.Namespace), metadataReader.GetString(reference.Name));
    }

    public string GetTypeFromSpecification(MetadataReader metadataReader, object? genericContext, TypeSpecificationHandle handle, byte rawTypeKind)
    {
        return metadataReader.GetTypeSpecification(handle).DecodeSignature(this, genericContext);
    }

    public string GetSystemType()
    {
        return "System.Type";
    }

    public string GetTypeFromSerializedName(string name)
    {
        return name;
    }

    public PrimitiveTypeCode GetUnderlyingEnumType(string type)
    {
        return PrimitiveTypeCode.Int32;
    }

    public bool IsSystemType(string type)
    {
        return type == "System.Type";
    }

    private static string Qualify(string namespaceName, string name)
    {
        if (namespaceName.Length == 0)
        {
            return name;
        }

        return namespaceName + "." + name;
    }
}

static class Program
{
    public static int Main(string[] args)
    {
        if (args.Length != 3)
        {
            Console.Error.WriteLine("usage: winmd-exporter <Windows.Win32.winmd> <apidocs.msgpack> <output.json>");

            return 2;
        }

        try
        {
            Export(args[0], args[1], args[2]);

            return 0;
        }
        catch (Exception exception)
        {
            Console.Error.WriteLine(exception);

            return 1;
        }
    }

    private static void Export(string inputPath, string docsPath, string outputPath)
    {
        using var docsStream = File.OpenRead(docsPath);
        var docs = MessagePackSerializer.Deserialize<Dictionary<string, ApiDetails>>(docsStream);
        var fieldComments = BuildFieldComments(docs.Values);

        using var stream = File.OpenRead(inputPath);
        using var peReader = new PEReader(stream);

        var reader = peReader.GetMetadataReader();
        var provider = new TypeNameProvider();
        var constants = new List<MetadataConstant>();
        var constantMethods = new List<MetadataConstantMethod>();
        var initializers = new List<MetadataInitializer>();
        var guids = new List<MetadataGuid>();

        foreach (var typeHandle in reader.TypeDefinitions)
        {
            var type = reader.GetTypeDefinition(typeHandle);
            var namespaceName = reader.GetString(type.Namespace);
            var typeName = reader.GetString(type.Name);
            var isEnum = IsEnum(reader, type);
            var guid = ReadGuidAttribute(reader, type.GetCustomAttributes(), provider);
            if (guid is not null)
            {
                var kind = (type.Attributes & TypeAttributes.Interface) != 0
                    ? "interface"
                    : "class";

                guids.Add(new MetadataGuid(namespaceName, typeName, kind, guid));
            }

            foreach (var fieldHandle in type.GetFields())
            {
                var field = reader.GetFieldDefinition(fieldHandle);
                var fieldName = reader.GetString(field.Name);
                var managedType = field.DecodeSignature(provider, null);
                var defaultHandle = field.GetDefaultValue();

                if (!defaultHandle.IsNil)
                {
                    var (kind, value) = ReadConstant(reader, defaultHandle);
                    var nativeType = ReadStringAttribute(reader, field.GetCustomAttributes(), "NativeTypeNameAttribute");
                    var documentation = ReadStringAttribute(reader, field.GetCustomAttributes(), "DocumentationAttribute");
                    var comment = ReadFieldComment(docs, fieldComments, typeName, fieldName);

                    constants.Add(new MetadataConstant(namespaceName, typeName, fieldName, managedType, nativeType, documentation, comment, kind, value, isEnum));
                }

                var initializer = ReadStringAttribute(reader, field.GetCustomAttributes(), "ConstantAttribute");
                if (initializer is not null)
                {
                    initializers.Add(new MetadataInitializer(namespaceName, typeName, fieldName, managedType, initializer));
                }
            }

            foreach (var methodHandle in type.GetMethods())
            {
                var method = reader.GetMethodDefinition(methodHandle);
                var value = ReadStringAttribute(reader, method.GetCustomAttributes(), "ConstantAttribute");
                if (value is null)
                {
                    continue;
                }

                var methodName = reader.GetString(method.Name);
                var signature = method.DecodeSignature(provider, null);
                var documentation = ReadStringAttribute(reader, method.GetCustomAttributes(), "DocumentationAttribute");
                var comment = ReadApiDescription(docs, methodName);

                constantMethods.Add(new MetadataConstantMethod(
                    namespaceName,
                    typeName,
                    methodName,
                    signature.ReturnType,
                    documentation,
                    comment,
                    value));
            }
        }

        constants.Sort(CompareConstants);
        constantMethods.Sort(CompareConstantMethods);
        initializers.Sort(CompareInitializers);
        guids.Sort(CompareGuids);

        var export = new MetadataExport(constants, constantMethods, initializers, guids);
        var options = new JsonSerializerOptions
        {
            WriteIndented = false,
        };

        using var output = File.Create(outputPath);

        JsonSerializer.Serialize(output, export, options);
    }

    private static string? ReadFieldComment(
        IReadOnlyDictionary<string, ApiDetails> docs,
        IReadOnlyDictionary<string, string> fieldComments,
        string declaringType,
        string fieldName)
    {
        if (!docs.TryGetValue(declaringType, out var details))
        {
            return ReadApiDescription(docs, fieldName) ?? ReadIndexedFieldComment(fieldComments, fieldName);
        }

        if (!details.Fields.TryGetValue(fieldName, out var comment))
        {
            return ReadApiDescription(docs, fieldName) ?? ReadIndexedFieldComment(fieldComments, fieldName);
        }

        comment = comment.Trim();
        if (comment.Length == 0)
        {
            return null;
        }

        return comment;
    }

    private static Dictionary<string, string> BuildFieldComments(IEnumerable<ApiDetails> docs)
    {
        var comments = new Dictionary<string, string>(StringComparer.Ordinal);
        var ambiguous = new HashSet<string>(StringComparer.Ordinal);

        foreach (var details in docs)
        {
            foreach (var field in details.Fields)
            {
                var comment = field.Value.Trim();
                if (comment.Length == 0 || ambiguous.Contains(field.Key))
                {
                    continue;
                }

                if (comments.TryGetValue(field.Key, out var existing) && existing != comment)
                {
                    comments.Remove(field.Key);
                    ambiguous.Add(field.Key);

                    continue;
                }

                comments[field.Key] = comment;
            }
        }

        return comments;
    }

    private static string? ReadIndexedFieldComment(
        IReadOnlyDictionary<string, string> fieldComments,
        string fieldName)
    {
        return fieldComments.TryGetValue(fieldName, out var comment) ? comment : null;
    }

    private static string? ReadApiDescription(
        IReadOnlyDictionary<string, ApiDetails> docs,
        string apiName)
    {
        if (!docs.TryGetValue(apiName, out var details))
        {
            return null;
        }

        var description = details.Description?.Trim();

        return description?.Length > 0 ? description : null;
    }

    private static int CompareConstants(MetadataConstant left, MetadataConstant right)
    {
        var result = string.CompareOrdinal(left.Namespace, right.Namespace);
        if (result != 0)
        {
            return result;
        }

        result = string.CompareOrdinal(left.DeclaringType, right.DeclaringType);
        if (result != 0)
        {
            return result;
        }

        return string.CompareOrdinal(left.Name, right.Name);
    }

    private static int CompareConstantMethods(MetadataConstantMethod left, MetadataConstantMethod right)
    {
        var result = string.CompareOrdinal(left.Namespace, right.Namespace);
        if (result != 0)
        {
            return result;
        }

        result = string.CompareOrdinal(left.DeclaringType, right.DeclaringType);
        if (result != 0)
        {
            return result;
        }

        return string.CompareOrdinal(left.Name, right.Name);
    }

    private static int CompareInitializers(MetadataInitializer left, MetadataInitializer right)
    {
        var result = string.CompareOrdinal(left.Namespace, right.Namespace);
        if (result != 0)
        {
            return result;
        }

        result = string.CompareOrdinal(left.DeclaringType, right.DeclaringType);
        if (result != 0)
        {
            return result;
        }

        return string.CompareOrdinal(left.Name, right.Name);
    }

    private static int CompareGuids(MetadataGuid left, MetadataGuid right)
    {
        var result = string.CompareOrdinal(left.Namespace, right.Namespace);
        if (result != 0)
        {
            return result;
        }

        return string.CompareOrdinal(left.Name, right.Name);
    }

    private static bool IsEnum(MetadataReader reader, TypeDefinition type)
    {
        var baseType = type.BaseType;
        if (baseType.Kind != HandleKind.TypeReference)
        {
            return false;
        }

        var reference = reader.GetTypeReference((TypeReferenceHandle)baseType);

        return reader.GetString(reference.Namespace) == "System" && reader.GetString(reference.Name) == "Enum";
    }

    private static (string Kind, string Value) ReadConstant(MetadataReader reader, ConstantHandle handle)
    {
        var constant = reader.GetConstant(handle);
        var blob = reader.GetBlobReader(constant.Value);

        return constant.TypeCode switch
        {
            ConstantTypeCode.Boolean => ("bool", (blob.ReadByte() != 0).ToString().ToLowerInvariant()),
            ConstantTypeCode.Char => ("char", ((uint)blob.ReadUInt16()).ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.SByte => ("int8", blob.ReadSByte().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.Byte => ("uint8", blob.ReadByte().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.Int16 => ("int16", blob.ReadInt16().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.UInt16 => ("uint16", blob.ReadUInt16().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.Int32 => ("int32", blob.ReadInt32().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.UInt32 => ("uint32", blob.ReadUInt32().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.Int64 => ("int64", blob.ReadInt64().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.UInt64 => ("uint64", blob.ReadUInt64().ToString(CultureInfo.InvariantCulture)),
            ConstantTypeCode.Single => ("float32", blob.ReadSingle().ToString("R", CultureInfo.InvariantCulture)),
            ConstantTypeCode.Double => ("float64", blob.ReadDouble().ToString("R", CultureInfo.InvariantCulture)),
            ConstantTypeCode.String => ("string", blob.ReadUTF16(blob.Length)),
            ConstantTypeCode.NullReference => ("null", ""),
            _ => throw new InvalidDataException($"Unsupported constant type {constant.TypeCode}"),
        };
    }

    private static string? ReadStringAttribute(
        MetadataReader reader,
        CustomAttributeHandleCollection handles,
        string attributeName)
    {
        foreach (var handle in handles)
        {
            var attribute = reader.GetCustomAttribute(handle);
            if (GetAttributeTypeName(reader, attribute) != attributeName)
            {
                continue;
            }

            var blob = reader.GetBlobReader(attribute.Value);
            if (blob.ReadUInt16() != 1)
            {
                throw new InvalidDataException($"Invalid {attributeName} prolog");
            }

            return blob.ReadSerializedString();
        }

        return null;
    }

    private static string? ReadGuidAttribute(
        MetadataReader reader,
        CustomAttributeHandleCollection handles,
        TypeNameProvider provider)
    {
        foreach (var handle in handles)
        {
            var attribute = reader.GetCustomAttribute(handle);
            if (GetAttributeTypeName(reader, attribute) != "GuidAttribute")
            {
                continue;
            }

            var arguments = attribute.DecodeValue(provider).FixedArguments;
            if (arguments.Length == 1 && arguments[0].Value is string text)
            {
                return Guid.Parse(text).ToString("D", CultureInfo.InvariantCulture);
            }

            if (arguments.Length == 11)
            {
                var guid = new Guid(
                    unchecked((int)Convert.ToUInt32(arguments[0].Value, CultureInfo.InvariantCulture)),
                    unchecked((short)Convert.ToUInt16(arguments[1].Value, CultureInfo.InvariantCulture)),
                    unchecked((short)Convert.ToUInt16(arguments[2].Value, CultureInfo.InvariantCulture)),
                    Convert.ToByte(arguments[3].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[4].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[5].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[6].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[7].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[8].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[9].Value, CultureInfo.InvariantCulture),
                    Convert.ToByte(arguments[10].Value, CultureInfo.InvariantCulture));

                return guid.ToString("D", CultureInfo.InvariantCulture);
            }

            throw new InvalidDataException("Unsupported GuidAttribute constructor");
        }

        return null;
    }

    private static string GetAttributeTypeName(MetadataReader reader, CustomAttribute attribute)
    {
        EntityHandle parent = attribute.Constructor.Kind switch
        {
            HandleKind.MemberReference => reader.GetMemberReference((MemberReferenceHandle)attribute.Constructor).Parent,
            HandleKind.MethodDefinition => reader.GetMethodDefinition((MethodDefinitionHandle)attribute.Constructor).GetDeclaringType(),
            _ => default,
        };

        return parent.Kind switch
        {
            HandleKind.TypeDefinition => reader.GetString(reader.GetTypeDefinition((TypeDefinitionHandle)parent).Name),
            HandleKind.TypeReference => reader.GetString(reader.GetTypeReference((TypeReferenceHandle)parent).Name),
            _ => "",
        };
    }
}
