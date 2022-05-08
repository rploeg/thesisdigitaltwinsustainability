using System;

namespace Company.Models
{
    public class Message
    {
        public string applicationId { get; set; }
        public string messageSource { get; set; }
        public string messageType { get; set; }
        public string deviceId { get; set; }
        public string schema { get; set; }
        public string templateId { get; set; }
        public DateTime enqueuedTime { get; set; }
        public Property[] properties { get; set; }
        public Enrichments enrichments { get; set; }
    }

    public class Enrichments
    {
        public string name { get; set; }
        public bool value { get; set; }
    }

    public class Property
    {
        public string name { get; set; }
        public bool value { get; set; }
    }
}